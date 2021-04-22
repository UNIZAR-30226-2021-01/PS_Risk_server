package baseDatos

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

const numTerritorios = 42
const maxMensajes = 100

/*
	Territorio almacena los datos de un territorio.
*/
type Territorio struct {
	IdTerritorio int `mapstructure:"id" json:"id"`
	IdJugador    int `mapstructure:"jugador" json:"jugador"`
	NumTropas    int `mapstructure:"tropas" json:"tropas"`
}

/*
	Jugador almacena datos reducidos de un usuario.
*/
type Jugador struct {
	Id             int    `mapstructure:"id" json:"id"`
	Nombre         string `mapstructure:"nombre" json:"nombre"`
	Icono          int    `mapstructure:"icono" json:"icono"`
	Aspecto        int    `mapstructure:"aspecto" json:"aspecto"`
	SigueVivo      bool   `mapstructure:"sigueVivo" json:"sigueVivo"`
	Refuerzos      int    `mapstructure:"refuerzos" json:"refuerzos"`
	NumTerritorios int    `mapstructure:"-" json:"numTerritorios"`
}

/*
	CrearJugador crea un jugador mediante los datos de un usuario.
*/
func CrearJugador(u Usuario) Jugador {
	return Jugador{
		Id:             u.Id,
		Nombre:         u.Nombre,
		Icono:          u.Icono,
		Aspecto:        u.Aspecto,
		SigueVivo:      true,
		Refuerzos:      0,
		NumTerritorios: 0,
	}
}

/*
	Partida almacena los datos relativos a una partida. Una partida sin iniciar es una
	sala de espera.

	Las etiquetas `mapstructure` son para codificar los datos que se envian a través de
	los websockets.

	Las etiquetas `json` son para codificar los datos que se guardan en la base de datos.
*/
type Partida struct {
	IdPartida    int                                  `mapstructure:"idPartida" json:"idPartida"`
	IdCreador    int                                  `mapstructure:"-" json:"idCreador"`
	TiempoTurno  int                                  `mapstructure:"tiempoTurno" json:"tiempoTurno"`
	TurnoActual  int                                  `mapstructure:"turnoActual,omitempty" json:"turnoActual"`
	TurnoJugador int                                  `mapstructure:"turnoJugador" json:"turnoJugador"`
	Fase         int                                  `mapstructure:"fase,omitempty" json:"fase"`
	Nombre       string                               `mapstructure:"nombrePartida" json:"nombrePartida"`
	Empezada     bool                                 `mapstructure:"-" json:"empezada"`
	Territorios  []Territorio                         `mapstructure:"territorios,omitempty" json:"territorios"`
	Jugadores    []Jugador                            `mapstructure:"jugadores" json:"jugadores"`
	Conexiones   sync.Map                             `mapstructure:"-" json:"-"`
	Mensajes     chan mensajesInternos.MensajePartida `mapstructure:"-" json:"-"`
	UltimoTurno  string                               `mapstructure:"ultimoTurno" json:"-"`
}

// Valores que puede tomar el campo Fase
const (
	faseRefuerzo   = iota + 1
	faseAtaque     = iota + 1
	faseMovimiento = iota + 1
)

/*
	Inicia la partida, devuelve error en caso de no poder hacerlo.
*/
func (p *Partida) IniciarPartida(idUsuario int) error {
	// Comprobar si se puede iniciar
	if p.Empezada {
		return errors.New("la partida ya está empezada")
	}
	if idUsuario != p.IdCreador {
		return errors.New("solo el creador de la partida puede comenzarla")
	}
	if len(p.Jugadores) < 3 {
		return errors.New("número de jugadores insuficiente")
	}

	// Decidir asignación de territorios
	p.asignarTerritorios()

	p.TurnoActual = 1
	p.Fase = faseRefuerzo
	p.Empezada = true
	p.UltimoTurno = time.Now().UTC().String()
	return nil
}

/*
	AnularInicio anula el inicio de una partida.
*/
func (p *Partida) AnularInicio() {
	p.Empezada = false
	p.Territorios = []Territorio{}
}

/*
	EntrarPartida añade un usuario a la partida.
	Devuelve error en caso de no poder hacerlo.
*/
func (p *Partida) EntrarPartida(u Usuario, ws *websocket.Conn) error {
	// Comprobar si se puede añadir
	if len(p.Jugadores) >= 6 {
		return errors.New("ya se ha alcanzado el número máximo de jugadores permitido")
	}
	if p.Empezada {
		return errors.New("no puedes unirte a una partida que ya ha empezado")
	}
	if p.EstaEnPartida(u.Id) {
		return errors.New("no puedes unirte a una partida en la que ya estás")
	}

	// Añadir el jugador
	p.Jugadores = append(p.Jugadores, CrearJugador(u))
	p.Conexiones.Store(u.Id, ws)
	return nil
}

// REQUIERE REVISION
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

/*
	EstaEnPartida devuelve si un jugador se encuentra en la partida o no.
*/
func (p *Partida) EstaEnPartida(idUsuario int) bool {
	for _, jugador := range p.Jugadores {
		if jugador.Id == idUsuario {
			return true
		}
	}
	return false
}

/*
	asignarTerritorios reparte todos los territorios del mapa entre los jugadores
	y distribuye las tropas.
*/
func (p *Partida) asignarTerritorios() {
	numJugadores := len(p.Jugadores)
	tropasTerritorio := tropasPorTerritorio(numJugadores)
	p.Territorios = make([]Territorio, numTerritorios)
	// Dar territorios aleatoriamente y colocar tropas en ellos
	ordenAsignacion := rand.Perm(numTerritorios)
	for i := 0; i < numTerritorios; i++ {
		idTerritorio := ordenAsignacion[i]
		p.Territorios[idTerritorio].IdTerritorio = idTerritorio
		p.Territorios[idTerritorio].IdJugador = i % numJugadores
		p.Territorios[idTerritorio].NumTropas = tropasTerritorio[i%numJugadores][i/numJugadores]
	}
	for i := range p.Jugadores {
		p.Jugadores[i].NumTerritorios = len(tropasTerritorio[i])
	}
}

/*
	tropasPorTerritorio devuelve un slice de numJugadores slices.
	El slice i-ésimo contiene cuántas tropas colocar en cada territorio del
	jugador i.
*/
func tropasPorTerritorio(numJugadores int) [][]int {
	var numTropasIniciales int
	switch numJugadores {
	case 3:
		numTropasIniciales = 35
	case 4:
		numTropasIniciales = 30
	case 5:
		numTropasIniciales = 25
	case 6:
		numTropasIniciales = 20
	}
	tropasTerritorio := make([][]int, numJugadores)
	for i := 0; i < numJugadores; i++ {
		if i < numTerritorios%numJugadores {
			tropasTerritorio[i] = make([]int, numTerritorios/numJugadores+1)
		} else {
			tropasTerritorio[i] = make([]int, numTerritorios/numJugadores)
		}
		for j := range tropasTerritorio[i] {
			tropasTerritorio[i][j] = numTropasIniciales / len(tropasTerritorio)
		}
	}
	return tropasTerritorio
}

/*
	Refuerzo coloca tropas de un jugador en un territorio y devuelve un JSON
	con el estado en el que ha quedado el territorio y un campo "_tipoMensaje".
	Si ocurre algún error lo devuelve en formato JSON.
*/
func (p *Partida) Refuerzo(idDestino, idJugador, refuerzos int) mensajes.JsonData {
	// Comprobar que la fase es correcta
	if p.Fase != faseRefuerzo {
		return mensajes.ErrorJsonPartida("No estás en la fase de refuerzo", 1)
	}
	// Comprobar que es el turno del jugador
	if p.TurnoJugador != idJugador {
		return mensajes.ErrorJsonPartida("No es tu turno", 1)
	}
	// Comprobar que el territorio pertenece al jugador
	if p.Territorios[idDestino].IdJugador != idJugador {
		return mensajes.ErrorJsonPartida("El territorio no pertenece a este jugador", 1)
	}
	// Comprobar que tiene suficientes refuerzos
	if p.Jugadores[idJugador].Refuerzos < refuerzos {
		return mensajes.ErrorJsonPartida("Se están intentando asignar más tropas"+
			" que las disponibles", 1)
	}
	if refuerzos < 0 {
		return mensajes.ErrorJsonPartida("No se puede asignar un número negativo"+
			"de tropas a un territorio", 1)
	}
	// Algoritmo de refuerzo
	p.Territorios[idDestino].NumTropas += refuerzos
	p.Jugadores[idJugador].Refuerzos -= refuerzos
	territorio := mensajes.JsonData{}
	mapstructure.Decode(p.Territorios[idDestino], &territorio)
	return mensajes.JsonData{
		"_tipoMensaje": "r",
		"territorio":   territorio,
	}
}

/*
	Ataque realiza un ataque entre los territorios indicados y devuelve un JSON
	con el estado de los territorios involucrados después del ataque y los
	valores obtenidos en los dados.
	Si ocurre algún error lo devuelve en formato JSON.
*/
func (p *Partida) Ataque(idOrigen, idDestino, idJugador, atacantes int) mensajes.JsonData {
	// Comprobar que la fase es correcta
	if p.Fase != faseAtaque {
		return mensajes.ErrorJsonPartida("No estás en la fase de ataque", 1)
	}
	// Comprobar que es el turno del jugador
	if p.TurnoJugador != idJugador {
		return mensajes.ErrorJsonPartida("No es tu turno", 1)
	}
	// Comprobar que el territorio del que parte el ataque pertenece al jugador
	if p.Territorios[idOrigen].IdJugador != idJugador {
		return mensajes.ErrorJsonPartida("No se puede atacar desde un territorio"+
			" que no te pertenece", 1)
	}
	// Comprobar que el territorio al que ataca no pertenece al jugador
	if p.Territorios[idDestino].IdJugador == idJugador {
		return mensajes.ErrorJsonPartida("No se puede atacar a un territorio"+
			"que ya te pertenece", 1)
	}
	// Comprobar que los territorios son adyacentes
	// TODO
	// Comprobar que se tienen suficientes tropas (y el origen no queda vacío)
	if p.Territorios[idOrigen].NumTropas <= atacantes {
		return mensajes.ErrorJsonPartida("No tienes tropas suficientes, siempre"+
			" debe quedar al menos una tropa en el territorio de origen", 1)
	}

	// Algoritmo ataque
	dadosAtaque := []int{}
	dadosDefensa := []int{}
	// Lanzar los dados del atacante
	for i := 0; i < 3 && i < atacantes; i++ {
		dadosAtaque[i] = rand.Intn(6) + 1
	}
	// Lanzar los dados del defensor
	defensores := min(3, p.Territorios[idDestino].NumTropas)
	for i := 0; i < defensores; i++ {
		dadosDefensa[i] = rand.Intn(6) + 1
	}
	// Ordenar los dados de mayor a menor
	sort.Sort(sort.Reverse(sort.IntSlice(dadosAtaque)))
	sort.Sort(sort.Reverse(sort.IntSlice(dadosDefensa)))
	// Resolver ataque
	for i := 0; i < min(len(dadosAtaque), len(dadosDefensa)); i++ {
		if dadosAtaque[i] > dadosDefensa[i] {
			// Gana atacante
			p.Territorios[idDestino].NumTropas--
		} else {
			// Gana defensor
			atacantes--
			p.Territorios[idOrigen].NumTropas--
		}
	}
	if p.Territorios[idDestino].NumTropas == 0 {
		// Gana atacante -> Mover tropas atacantes supervivientes
		idDefensor := p.Territorios[idDestino].IdJugador
		p.Territorios[idDestino].IdJugador = idJugador
		p.Territorios[idDestino].NumTropas += atacantes
		p.Territorios[idOrigen].NumTropas -= atacantes
		p.Jugadores[idJugador].NumTerritorios++
		p.Jugadores[idDefensor].NumTerritorios--
		if p.Jugadores[idDefensor].NumTerritorios == 0 {
			p.Jugadores[idDefensor].SigueVivo = false
		}
	}
	territorioOrigen := Territorio{}
	territorioDestino := Territorio{}
	mapstructure.Decode(p.Territorios[idOrigen], &territorioOrigen)
	mapstructure.Decode(p.Territorios[idDestino], &territorioDestino)
	return mensajes.JsonData{
		"_tipoMensaje":      "a",
		"territorioOrigen":  territorioOrigen,
		"territorioDestino": territorioDestino,
		"dadosOrigen":       dadosAtaque,
		"dadosDestrino":     dadosDefensa,
	}
}

func (p *Partida) Movimiento(idOrigen, idDestino, idJugador, tropas int) mensajes.JsonData {
	// Comprobar que la fase es correcta
	if p.Fase != faseMovimiento {
		return mensajes.ErrorJsonPartida("No estás en la fase de movimiento", 1)
	}
	// Comprobar que es el turno del jugador
	if p.TurnoJugador != idJugador {
		return mensajes.ErrorJsonPartida("No es tu turno", 1)
	}
	// Comprobar que existe ruta entre territorios del jugador
	// TODO
	// Comprobar que se tienen suficientes tropas (y el origen no queda vacío)
	if p.Territorios[idOrigen].NumTropas <= tropas {
		return mensajes.ErrorJsonPartida("No tienes tropas suficientes, siempre"+
			" debe quedar al menos una tropa en el territorio de origen", 1)
	}
	// Mover las tropas
	p.Territorios[idOrigen].NumTropas -= tropas
	p.Territorios[idDestino].NumTropas += tropas
	// Codificar los datos en formato json
	territorioOrigen := Territorio{}
	territorioDestino := Territorio{}
	mapstructure.Decode(p.Territorios[idOrigen], &territorioOrigen)
	mapstructure.Decode(p.Territorios[idDestino], &territorioDestino)
	return mensajes.JsonData{
		"_tipoMensaje":      "m",
		"territorioOrigen":  territorioOrigen,
		"territorioDestino": territorioDestino,
	}
}

func (p *Partida) AsignarRefuerzos() {
	for i := range p.Jugadores {
		// Dar a todos los jugadores el numero de refuerzos que les corresponde
		// por número de territorios
		p.Jugadores[i].Refuerzos = p.Jugadores[i].NumTerritorios / 3
		if p.Jugadores[i].Refuerzos < 3 {
			p.Jugadores[i].Refuerzos = 3
		}
	}

	// TODO REFUERZOS POR CONTINENTES

	// TODO SUSTITO A REFUERZOS POR CARTA
}

func (p *Partida) ObtenerPosicionJugador(id int) int {
	for i := range p.Jugadores {
		if p.Jugadores[i].Id == id {
			return i
		}
	}
	return -1
}

func (p *Partida) AvanzarFase(jugador int) mensajes.JsonData {

	if p.TurnoJugador != jugador {
		return mensajes.ErrorJsonPartida("No es tu turno", 1)
	}

	res := mensajes.JsonData{"_tipoMensaje": "f"}

	switch p.Fase {
	case faseRefuerzo:
		//if p.Jugadores[jugador].Refuerzos > 0 {
		//	return mensajes.ErrorJsonPartida("Aún te quedan refuerzos", 1)
		//}
		p.Fase++
		return res
	case faseAtaque:
		p.Fase++
		return res
	case faseMovimiento:
		p.Fase = faseRefuerzo
		p.TurnoActual++
		// Calcular el jugador de cual es el turno
		p.TurnoJugador = p.TurnoActual % len(p.Jugadores)
		for !p.Jugadores[p.TurnoJugador].SigueVivo {
			p.TurnoActual++
			p.TurnoJugador = p.TurnoActual % len(p.Jugadores)
		}
		// Calcular el nuevo valor de los refuerzos
		p.AsignarRefuerzos()
		// Nueva marca temporal del ultimo turno
		p.UltimoTurno = time.Now().UTC().String()
		// Codificar los datos de la partida en formato json
		mapstructure.Decode(p, &res)
		res["_tipoMensaje"] = "p"
		return res
	}

	return mensajes.ErrorJsonPartida("La partida no está empezada", 1)

}

// Funciones de envío de mensaje a traves de WebSockets

func (p *Partida) EnviarATodos(mensaje mensajes.JsonData) {
	for _, jugador := range p.Jugadores {
		ws, ok := p.Conexiones.Load(jugador.Id)
		if ok {
			ws.(*websocket.Conn).WriteJSON(mensaje)
		}
	}
}

func (p *Partida) EnviarError(idUsuario, code int, err string) {
	ws, ok := p.Conexiones.Load(idUsuario)
	if ok {
		ws.(*websocket.Conn).WriteJSON(mensajes.ErrorJsonPartida(err, code))
	}
}

func (p *Partida) Enviar(id int, mensaje mensajes.JsonData) {
	ws, ok := p.Conexiones.Load(id)
	if ok {
		ws.(*websocket.Conn).WriteJSON(mensaje)
	}
}

// FUNCIONES AUXILIARES PARA EL MANEJO DE ARRAYS

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
