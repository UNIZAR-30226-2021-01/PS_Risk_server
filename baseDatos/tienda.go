package baseDatos

import "PS_Risk_server/mensajes"

/*
	Tienda contiene iconos y aspectos en formato json.
*/
type Tienda struct {
	Iconos, Aspectos []mensajes.JsonData
}

/*
	ObtenerPrecioIcono devuelve el precio de un icono. Si no se he encuentra el icono
	devuelve falso, si lo encuentra verdadero.
*/
func (t *Tienda) ObtenerPrecioIcono(id int) (int, bool) {
	for _, icono := range t.Iconos {
		if icono["id"].(int) == id {
			return icono["precio"].(int), true
		}
	}
	return 0, false
}

/*
	ObtenerPrecioAspecto devuelve el precio de un aspecto. Si no se he encuentra el
	aspecto devuelve falso, si lo encuentra verdadero.
*/
func (t *Tienda) ObtenerPrecioAspecto(id int) (int, bool) {
	for _, aspecto := range t.Aspectos {
		if aspecto["id"].(int) == id {
			return aspecto["precio"].(int), true
		}
	}
	return 0, false
}
