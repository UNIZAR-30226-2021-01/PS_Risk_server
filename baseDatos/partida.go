package baseDatos

import (
	"PS_Risk_server/mensajesInternos"
	"errors"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

const numTerritorios = 42
const MaxMensajes = 100

type Territorio struct {
	IdJugador int `mapstructure:"idJugador" json:"idJugador"`
	NumTropas int `mapstructure:"numTropas" json:"numTropas"`
}

type Jugador struct {
	Id        int    `mapstructure:"id" json:"id"`
	Nombre    string `mapstructure:"nombre" json:"nombre"`
	Icono     int    `mapstructure:"icono" json:"icono"`
	Aspecto   int    `mapstructure:"aspecto" json:"aspecto"`
	SigueVivo bool   `mapstructure:"sigueVivo" json:"sigueVivo"`
}

func crearJugador(u Usuario) Jugador {
	return Jugador{
		Id:        u.Id,
		Nombre:    u.Nombre,
		Icono:     u.Icono,
		Aspecto:   u.Aspecto,
		SigueVivo: true,
	}
}

type Partida struct {
	IdPartida   int                                  `mapstructure:"idPartida" json:"idPartida"`
	IdCreador   int                                  `mapstructure:"-" json:"idCreador"`
	TiempoTurno int                                  `mapstructure:"tiempoTurno" json:"tiempoTurno"`
	TurnoActual int                                  `mapstructure:"turnoActual,omitempty" json:"turnoActual"`
	Fase        int                                  `mapstructure:"fase,omitempty" json:"fase"`
	Nombre      string                               `mapstructure:"nombrePartida" json:"nombrePartida"`
	Empezada    bool                                 `mapstructure:"-" json:"empezada"`
	Territorios []Territorio                         `mapstructure:"territorios,omitempty" json:"territorios"`
	Jugadores   []Jugador                            `mapstructure:"jugadores" json:"jugadores"`
	Conexiones  sync.Map                             `mapstructure:"-" json:"-"`
	Mensajes    chan mensajesInternos.MensajePartida `mapstructure:"-" json:"-"`
}

func (p *Partida) IniciarPartida(idUsuario int) error {
	if p.Empezada {
		return errors.New("la partida ya está empezada")
	}
	if idUsuario != p.IdCreador {
		return errors.New("solo el creador de la partida puede comenzarla")
	}
	if len(p.Jugadores) < 3 {
		return errors.New("número de jugadores insuficiente")
	}
	// Decidir orden de jugadores
	orden := rand.Perm(len(p.Jugadores))
	aux := make([]Jugador, len(p.Jugadores))
	copiados := copy(aux, p.Jugadores)
	if copiados != len(p.Jugadores) {
		return errors.New("error al copiar los jugadores a un vector auxiliar")
	}
	for i := 0; i < len(p.Jugadores); i++ {
		p.Jugadores[orden[i]] = aux[i]
	}
	// Decidir asignación de territorios
	p.Territorios = make([]Territorio, 0, numTerritorios)
	t := Territorio{
		IdJugador: 0,
		NumTropas: 0,
	}
	for i := 0; i < numTerritorios; i++ {
		p.Territorios = append(p.Territorios, t)
	}
	asignados := make([]int, len(p.Jugadores))
	for i := 0; i < len(p.Jugadores); {
		idTerritorio := rand.Intn(numTerritorios)
		if !contenido(asignados, idTerritorio) {
			asignados[i] = idTerritorio
			p.Territorios[idTerritorio].IdJugador = p.Jugadores[i].Id
			i++
		}
	}
	p.Empezada = true
	// Faltan parametros de la partida empezada por asignar
	return nil
}

func (p *Partida) AnularInicio() {
	p.Empezada = false
	p.Territorios = []Territorio{}
}

func (p *Partida) EntrarPartida(u Usuario, ws *websocket.Conn) error {
	if len(p.Jugadores) >= 6 {
		return errors.New("ya se ha alcanzado el número máximo de jugadores permitido")
	}
	if p.Empezada {
		return errors.New("no se puede unir a una partida que ya ha empezado")
	}
	if p.EstaEnPartida(u.Id) {
		return errors.New("no puedes unirte a una partida en la que ya estás")
	}
	p.Jugadores = append(p.Jugadores, crearJugador(u))
	p.Conexiones.Store(u.Id, ws)
	return nil
}

func (p *Partida) ExpulsarDePartida(idJugador int) error {
	if !p.EstaEnPartida(idJugador) {
		return errors.New("el jugador no está en la partida, no se puede retirar")
	}
	if p.Empezada {
		for _, jugador := range p.Jugadores {
			if jugador.Id == idJugador {
				jugador.SigueVivo = false
				break
			}
		}
		p.Conexiones.Delete(idJugador)
	} else if p.IdCreador == idJugador {
		return errors.New("no se puede eliminar de la partida al creador")
	} else {
		p.Jugadores = borrar(p.Jugadores, idJugador)
		p.Conexiones.Delete(idJugador)
	}
	return nil
}

func contenido(lista []int, valor int) bool {
	i := indice(lista, valor)
	return i >= 0
}

func indice(lista []int, valor int) int {
	for i, v := range lista {
		if v == valor {
			return i
		}
	}
	return -1
}

func borrar(lista []Jugador, valor int) []Jugador {
	i := -1
	for ind, v := range lista {
		if v.Id == valor {
			i = ind
			break
		}
	}
	if i < 0 {
		return lista
	} else {
		return append(lista[:i], lista[i+1:]...)
	}
}

func (p *Partida) EstaEnPartida(idUsuario int) bool {
	for _, jugador := range p.Jugadores {
		if jugador.Id == idUsuario {
			return true
		}
	}
	return false
}
