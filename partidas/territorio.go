package partidas

import "PS_Risk_server/mensajes"

const numTerritorios = 42

type Territorio struct {
	IdJugador, NumTropas int
}

func (t *Territorio) ToJSON() mensajes.JsonData {
	return mensajes.JsonData{
		"numJugador": t.IdJugador,
		"tropas":     t.NumTropas,
	}
}
