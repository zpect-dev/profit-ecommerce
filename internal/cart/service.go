package cart

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Service define el orquestador de lógica de negocio del carrito.
type Service interface {
	AddToCart(ctx context.Context, userID string, item CartItem) error
	GetValidatedCart(ctx context.Context, userID string) (*Cart, error)
	Close() error
}

type cartService struct {
	cacheRepo      CartCacheRepository
	dbRepo         CartDBRepository
	catalogService CatalogService

	// Canal para procesar las operaciones Write-Behind a la BD
	persistCh chan Cart
	wg        sync.WaitGroup
}

// NewService crea una instancia del servicio del carrito e inicia
// el worker en segundo plano para manejar el patrón Write-Behind.
func NewService(cacheRepo CartCacheRepository, dbRepo CartDBRepository, catalogService CatalogService) Service {
	s := &cartService{
		cacheRepo:      cacheRepo,
		dbRepo:         dbRepo,
		catalogService: catalogService,
		persistCh:      make(chan Cart, 1000), // Buffer para evitar bloquear la API HTTP
	}

	s.wg.Add(1)
	go s.worker()

	return s
}

// AddToCart agrega un item, actualiza el caché de Redis síncronamente y envía un evento
// por canal para persistir en BD asíncronamente (Write-Behind).
func (s *cartService) AddToCart(ctx context.Context, userID string, item CartItem) error {
	// 1. Obtener carrito actual desde caché
	cart, err := s.cacheRepo.GetCart(ctx, userID)
	if err != nil {
		if err.Error() == "redis: nil" {
			// El carrito no existe, iniciamos uno vacío
			cart = Cart{
				UserID:    userID,
				Items:     []CartItem{},
				UpdatedAt: time.Now(),
			}
		} else {
			return err
		}
	}

	// 2. Lógica del dominio: Actualizar o agregar item
	found := false
	for i, existingItem := range cart.Items {
		if existingItem.ProductID == item.ProductID {
			cart.Items[i].Quantity += item.Quantity
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, item)
	}
	cart.UpdatedAt = time.Now()

	// 3. Guardar en Caché síncronamente
	if err := s.cacheRepo.SaveCart(ctx, cart); err != nil {
		return err
	}

	// 4. Enviar al worker de la base de datos de forma segura
	select {
	case s.persistCh <- cart:
		// Evento enviado correctamente para su persistencia diferida
	default:
		// Si el canal está lleno, se registra el error sin bloquear el request síncrono HTTP
		slog.Error("AddToCart: buffer Write-Behind lleno, carrito no enviado a DB", "userID", userID)
	}

	return nil
}

// GetValidatedCart obtiene el carrito, verifica el stock real contra el Catálogo,
// ajusta según el inventario disponible y sincroniza con BD si ocurrieron quitas o cambios.
func (s *cartService) GetValidatedCart(ctx context.Context, userID string) (*Cart, error) {
	cart, err := s.cacheRepo.GetCart(ctx, userID)
	if err != nil {
		if err.Error() == "redis: nil" {
			return &Cart{UserID: userID, Items: []CartItem{}, UpdatedAt: time.Now()}, nil
		}
		return nil, err
	}

	if len(cart.Items) == 0 {
		return &cart, nil
	}

	// 1. Array temporal de productIDs
	var pids []string
	for _, item := range cart.Items {
		pids = append(pids, item.ProductID)
	}

	// 2. Solicitar stock al dominio catálogo (HTTP real u otra fuente externa)
	stockMap, err := s.catalogService.CheckStock(ctx, pids)
	if err != nil {
		return nil, err
	}

	// 3. Evaluar y ajustar el carrito actual
	var validItems []CartItem
	modified := false

	for _, item := range cart.Items {
		stock, ok := stockMap[item.ProductID]
		if !ok || stock <= 0 {
			// Remover item del todo, el stock quedó en 0 o fue despublicado
			modified = true
			continue
		}
		if item.Quantity > stock {
			// Ajustar parcialmente a lo disponible
			item.Quantity = stock
			modified = true
		}
		validItems = append(validItems, item)
	}

	// 4. Si el carrito se alteró, sincronizar persistencia en Write-Behind
	if modified {
		cart.Items = validItems
		cart.UpdatedAt = time.Now()

		if err := s.cacheRepo.SaveCart(ctx, cart); err != nil {
			slog.Error("GetValidatedCart: fallo guardando cart modificado en caché", "err", err, "userID", userID)
			return nil, err
		}

		select {
		case s.persistCh <- cart:
		default:
			slog.Error("GetValidatedCart: buffer persistencia lleno, ajuste no guardado en DB", "userID", userID)
		}
	}

	return &cart, nil
}

// worker recibe los carritos a través del canal y persiste en BD garantizando que
// no hayan fugas de goroutines por cada llamada a AddToCart HTTP.
func (s *cartService) worker() {
	defer s.wg.Done()

	// El worker procesará todos los elementos pendientes en el buffer antes de salir
	for cart := range s.persistCh {
		// Usamos Background ya que el Context del Endpoint HTTP original puede haber sido cancelado (petición HTTP finalizada)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		if err := s.dbRepo.PersistCart(ctx, cart); err != nil {
			slog.Error("Fallo al persistir carrito en la Base de Datos (Write-Behind)",
				"error", err,
				"userID", cart.UserID)
		}

		cancel()
	}
}

// Close detiene el worker en forma Graceful esperando que finalice.
func (s *cartService) Close() error {
	close(s.persistCh) // Cerramos el canal para evitar nuevas inserciones y drenar el buffer
	s.wg.Wait()        // Esperamos a que procese todos los pendientes
	return nil
}
