package partidas

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"errors"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

const MaxMensajes = 100

type Partida struct {
	partidaJson
	IdCreador      int
	Jugadores      []int
	JugadoresVivos map[int]bool
	Conexiones     sync.Map
	Mensajes       chan mensajesInternos.MensajePartida
}

type partidaJson struct {
	IdPartida   int          `mapstructure:"idSala"`
	TiempoTurno int          `mapstructure:"tiempoTurno"`
	TurnoActual int          `mapstructure:"turnoActual,omitempty"`
	Fase        int          `mapstructure:"fase,omitempty"`
	Nombre      string       `mapstructure:"nombreSala"`
	Empezada    bool         `mapstructure:"empezada"`
	Territorios []Territorio `mapstructure:"listaTerritorios,omitempty"`
}

type jugadorJson struct {
	Id        int    `mapstructure:"id"`
	Nombre    string `mapstructure:"nombre,omitempty"`
	Icono     int    `mapstructure:"icono,omitempty"`
	Aspecto   int    `mapstructure:"aspecto,omitempty"`
	SigueVivo bool   `mapstructure:"sigueVivo"`
}

func NuevaPartida(idPartida, idCreador, tiempoTurno int, nombreSala string,
	wsCreador *websocket.Conn) *Partida {
	pJson := partidaJson{
		IdPartida:   idPartida,
		TiempoTurno: tiempoTurno,
		Nombre:      nombreSala,
		Empezada:    false,
	}
	p := &Partida{
		partidaJson:    pJson,
		IdCreador:      idCreador,
		Jugadores:      []int{idCreador},
		JugadoresVivos: make(map[int]bool),
		Conexiones:     sync.Map{},
		Mensajes:       make(chan mensajesInternos.MensajePartida, MaxMensajes),
	}
	p.Conexiones.Store(idCreador, wsCreador)
	return p
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
	aux := make([]int, len(p.Jugadores))
	copiados := copy(aux, p.Jugadores)
	if copiados != len(p.Jugadores) {
		return errors.New("error al copiar los jugadores a un vector auxiliar")
	}
	for i := 0; i < len(p.Jugadores); i++ {
		p.Jugadores[orden[i]] = aux[i]
	}
	// Marcar que están vivos
	for _, idJugador := range p.Jugadores {
		p.JugadoresVivos[idJugador] = true
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
			p.Territorios[idTerritorio].IdJugador = p.Jugadores[i]
			i++
		}
	}

	p.Empezada = true
	return nil
}

func (p *Partida) EntrarPartida(idUsuario int, ws *websocket.Conn) error {
	if len(p.Jugadores) >= 6 {
		return errors.New("ya se ha alcanzado el número máximo de jugadores permitido")
	}
	if p.Empezada {
		return errors.New("no se puede unir a una partida que ya ha empezado")
	}
	if contenido(p.Jugadores, idUsuario) {
		return errors.New("no puedes unirte a una partida en la que ya estás")
	}
	p.Jugadores = append(p.Jugadores, idUsuario)
	p.Conexiones.Store(idUsuario, ws)
	return nil
}

func (p *Partida) AbandonarPartida(idJugador int) error {
	if !p.EstaEnPartida(idJugador) {
		return errors.New("el jugador no está en la partida, no se puede retirar")
	}
	if p.Empezada {
		p.JugadoresVivos[idJugador] = false
		p.Conexiones.Delete(idJugador)
	} else if p.IdCreador == idJugador {
		return errors.New("no se puede eliminar de la partida al creador")
	} else {
		p.Jugadores = borrar(p.Jugadores, idJugador)
		p.Conexiones.Delete(idJugador)
	}
	return nil
}

func (p *Partida) EstaEnPartida(idUsuario int) bool {
	return contenido(p.Jugadores, idUsuario)
}

func PartidaDesdeJson(estado mensajes.JsonData, idCreador int) (*Partida, error) {
	if i, ok := estado["idSala"]; !ok || i == nil {
		return &Partida{}, errors.New("el json no contiene datos sobre una partida")
	}
	var pJson partidaJson
	err := mapstructure.Decode(estado, &pJson)
	if err != nil {
		return &Partida{}, err
	}
	jugadores := []int{}
	jugadoresVivos := make(map[int]bool)
	j := estado["jugadores"].([]interface{})
	var jugador jugadorJson
	for _, datosJugador := range j {
		err = mapstructure.Decode(datosJugador, &jugador)
		if err != nil {
			return &Partida{}, err
		}
		id := jugador.Id
		jugadores = append(jugadores, id)
		jugadoresVivos[id] = jugador.SigueVivo
	}
	p := &Partida{
		partidaJson:    pJson,
		IdCreador:      idCreador,
		Jugadores:      jugadores,
		JugadoresVivos: jugadoresVivos,
		Conexiones:     sync.Map{},
		Mensajes:       make(chan mensajesInternos.MensajePartida, MaxMensajes),
	}
	return p, nil
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

func borrar(lista []int, valor int) []int {
	i := indice(lista, valor)
	if i < 0 {
		return lista
	} else {
		return append(lista[:i], lista[i+1:]...)
	}
}
