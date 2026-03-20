package syncer

// LinArt representa una línea de artículo (tabla lin_art).
type LinArt struct {
	CoLin  string `db:"co_lin"`
	LinDes string `db:"lin_des"`
}

// CatArt representa una categoría de artículo (tabla cat_art).
type CatArt struct {
	CoCat  string `db:"co_cat"`
	CatDes string `db:"cat_des"`
}

// SubLin representa una sub-línea (tabla sub_lin).
type SubLin struct {
	CoSubl  string `db:"co_subl"`
	SublDes string `db:"subl_des"`
	CoLin   string `db:"co_lin"`
}

// Almacen representa un almacén (tabla almacen).
type Almacen struct {
	CoAlma  string `db:"co_alma"`
	AlmaDes string `db:"alma_des"`
}

// SubAlma representa un sub-almacén (tabla sub_alma).
type SubAlma struct {
	CoSub  string `db:"co_sub"`
	DesSub string `db:"des_sub"`
	CoAlma string `db:"co_alma"`
}

// Descuento representa un descuento (tabla descuen).
type Descuento struct {
	CoDesc   string  `db:"co_desc"`
	TipoCli  string  `db:"tipo_cli"`
	TipoDesc string  `db:"tipo_desc"`
	Porc1    float64 `db:"porc1"`
}

// Article representa un artículo leído desde Profit (tabla art).
// ImageURL es generado por el Service (lógica de negocio), no por el repositorio.
type Article struct {
	CoArt    string  `db:"co_art"`
	ArtDes   string  `db:"art_des"`
	StockAct float64 `db:"stock_act"`
	PrecVta1 float64 `db:"prec_vta1"`
	PrecVta2 float64 `db:"prec_vta2"`
	PrecVta3 float64 `db:"prec_vta3"`
	PrecVta4 float64 `db:"prec_vta4"`
	PrecVta5 float64 `db:"prec_vta5"`
	TipoImp  string  `db:"tipo_imp"`
	CoLin    string  `db:"co_lin"`
	CoCat    string  `db:"co_cat"`
	CoSubl   string  `db:"co_subl"`
	Campo4   string  `db:"campo4"`
	ImageURL string  // Generado por Service, no mapeado a BD origen
}

// StAlmac representa un registro de stock por almacén (tabla st_almac).
type StAlmac struct {
	CoAlma    string  `db:"co_alma"`
	CoArt     string  `db:"co_art"`
	StockAct  float64 `db:"stock_act"`
	SStockAct float64 `db:"sstock_act"`
	StockCom  float64 `db:"stock_com"`
	SStockCom float64 `db:"sstock_com"`
	StockLle  float64 `db:"stock_lle"`
	SStockLle float64 `db:"sstock_lle"`
	StockDes  float64 `db:"stock_des"`
	SStockDes float64 `db:"sstock_des"`
}

// TipoCli representa la tabla de de clasificaciones para el cliente (profit tip_cli).
type TipoCli struct {
	TipCli  string `db:"tip_cli" json:"tip_cli"`
	DesTipo string `db:"des_tipo" json:"des_tipo"`
	PrecioA string `db:"precio_a" json:"precio_a"`
}

// Cliente representa el registro master del cliente extraído desde (profit clientes).
type Cliente struct {
	CoCli    string  `db:"co_cli" json:"co_cli"`
	Tipo     string  `db:"tipo" json:"tipo"`
	CliDes   string  `db:"cli_des" json:"cli_des"`
	Rif      string  `db:"rif" json:"rif"`
	Inactivo bool    `db:"inactivo" json:"inactivo"`
	Login    float64 `db:"login" json:"login"`
	MontCre  float64 `db:"mont_cre" json:"mont_cre"`
}
