package catalog

// Category representa una línea de artículos (categoría principal).
type Category struct {
	Id     string `db:"co_lin" json:"id"`
	Titulo string `db:"lin_des" json:"titulo"`
}

// AlmacenResumen es el detalle de inventario por almacén, deserializado del JSONB.
type AlmacenResumen struct {
	Nombre            string  `json:"nombre"`
	StockTotal        float64 `json:"stock_total"`
	StockComprometido float64 `json:"stock_comprometido"`
	StockPorLlegar    float64 `json:"stock_por_llegar"`
}

// Product representa un artículo del catálogo con toda su información.
type Product struct {
	CoArt         string  `db:"co_art" json:"co_art"`
	ArtDes        string  `db:"art_des" json:"art_des"`
	StockAct      float64 `db:"stock_act" json:"stock_act"`
	PrecVta1      float64 `db:"prec_vta1" json:"prec_vta1"`
	PrecVta2      float64 `db:"prec_vta2" json:"prec_vta2"`
	PrecVta3      float64 `db:"prec_vta3" json:"prec_vta3"`
	PrecVta4      float64 `db:"prec_vta4" json:"prec_vta4"`
	PrecVta5      float64 `db:"prec_vta5" json:"prec_vta5"`
	TipoImp       string  `db:"tipo_imp" json:"tipo_imp"`
	CoLin         string  `db:"co_lin" json:"co_lin"`
	CoCat         string  `db:"co_cat" json:"co_cat"`
	CoSubl        string  `db:"co_subl" json:"co_subl"`
	ImageUrl      string  `db:"image_url" json:"image_url"`
	DescArticulo  float64 `db:"desc_articulo" json:"desc_articulo"`
	DescCategoria float64 `db:"desc_categoria" json:"desc_categoria"`
	DescLinea     float64 `db:"desc_linea" json:"desc_linea"`

	// InventarioRaw recibe el raw byte del SQL. El tag debe coincidir con el alias del query.
	InventarioRaw []byte                    `db:"inventario_detallado" json:"-"`
	Inventario    map[string]AlmacenResumen `db:"-" json:"inventario"`
}
