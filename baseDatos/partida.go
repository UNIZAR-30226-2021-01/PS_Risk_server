package baseDatos

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"errors"
	"math/rand"
	"net/smtp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jordan-wright/email"
	"github.com/mitchellh/mapstructure"
)

const (
	maxMensajes     = 100
	maxDadosDefensa = 2
	maxDadosAtaque  = 3
)

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
	Id             int    `json:"id"`
	Correo         string `json:"-"`
	Nombre         string `json:"nombre"`
	Icono          int    `json:"icono"`
	Aspecto        int    `json:"aspecto"`
	SigueVivo      bool   `json:"sigueVivo"`
	Refuerzos      int    `json:"refuerzos"`
	NumTerritorios int    `json:"numTerritorios"`
}

/*
	CrearJugador crea un jugador mediante los datos de un usuario.
*/
func CrearJugador(u Usuario) Jugador {
	var correo string

	if !u.RecibeCorreos {
		correo = ""
	} else {
		correo = u.Correo
	}

	return Jugador{
		Id:             u.Id,
		Correo:         correo,
		Nombre:         u.Nombre,
		Icono:          u.Icono,
		Aspecto:        u.Aspecto,
		SigueVivo:      true,
		Refuerzos:      0,
		NumTerritorios: 0,
	}
}

/*
	ActualizarJugador actualiza los datos del jugador indicado para que
	coincidan con los de u.
*/
func (j *Jugador) ActualizarJugador(u Usuario) {
	var correo string

	if !u.RecibeCorreos {
		correo = ""
	} else {
		correo = u.Correo
	}

	j.Nombre = u.Nombre
	j.Correo = correo
	j.Icono = u.Icono
	j.Aspecto = u.Aspecto
}

/*
	Mano almacena la cantidad de cartas de refuerzo de cada tipo que tiene un
	jugador.
*/
type Mano struct {
	Infanteria int `json:"infanteria"`
	Caballeria int `json:"caballeria"`
	Artilleria int `json:"artilleria"`
}

/*
	Informar devuelve un mensaje indicando cuántos refuerzos extra se consiguen
	por sets de cartas, o una cadena vacía si no se consiguen refuerzos extra.
*/
func (m *Mano) Informar() string {
	if m.Artilleria >= 1 && m.Caballeria >= 1 && m.Infanteria >= 1 {
		return "Has encontrado un gran botín, recibirás 10 refuerzos extra."
	}
	if m.Artilleria >= 3 {
		return "Has encontrado un botín, recibirás 8 refuerzos extra."
	}
	if m.Caballeria >= 3 {
		return "Has encontrado un pequeño botín, recibirás 6 refuerzos extra."
	}
	if m.Infanteria >= 3 {
		return "Has encontrado unos pocos recursos, recibirás 4 refuerzos extra."
	}
	return ""
}

/*
	Negociar intercambia 3 cartas de la mano por tropas de refuerzo extra,
	siempre de la forma que más tropas proporciona, y devuelve cuántas tropas
	se consiguen.
*/
func (m *Mano) Negociar() int {
	if m.Artilleria >= 1 && m.Caballeria >= 1 && m.Infanteria >= 1 {
		m.Infanteria--
		m.Caballeria--
		m.Artilleria--
		return 10
	}
	if m.Artilleria >= 3 {
		m.Artilleria -= 3
		return 8
	}
	if m.Caballeria >= 3 {
		m.Caballeria -= 3
		return 6
	}
	if m.Infanteria >= 3 {
		m.Infanteria -= 3
		return 4
	}
	return 0
}

/*
	Robar añade una carta del tipo indicado a la mano.
*/
func (m *Mano) Robar(c int) {
	switch c {
	case 0:
		m.Artilleria++
	case 1:
		m.Caballeria++
	case 2:
		m.Artilleria++
	}
}

/*
	Partida almacena los datos relativos a una partida. Una partida sin iniciar es una
	sala de espera.

	Las etiquetas `mapstructure` son para codificar los datos que se envían a través de
	los websockets.

	Las etiquetas `json` son para codificar los datos que se guardan en la base de datos.
*/
type Partida struct {
	IdPartida           int                                  `mapstructure:"idPartida" json:"idPartida"`
	IdCreador           int                                  `mapstructure:"-" json:"idCreador"`
	TiempoTurno         int                                  `mapstructure:"tiempoTurno" json:"tiempoTurno"`
	TurnoActual         int                                  `mapstructure:"turnoActual,omitempty" json:"turnoActual"`
	TurnoJugador        int                                  `mapstructure:"turnoJugador" json:"turnoJugador"`
	Fase                int                                  `mapstructure:"fase,omitempty" json:"fase"`
	Nombre              string                               `mapstructure:"nombrePartida" json:"nombrePartida"`
	Empezada            bool                                 `mapstructure:"-" json:"empezada"`
	Territorios         []Territorio                         `mapstructure:"territorios,omitempty" json:"territorios"`
	Jugadores           []Jugador                            `mapstructure:"jugadores" json:"jugadores"`
	Conexiones          sync.Map                             `mapstructure:"-" json:"-"`
	Mensajes            chan mensajesInternos.MensajePartida `mapstructure:"-" json:"-"`
	UltimoTurno         string                               `mapstructure:"ultimoTurno,omitempty" json:"-"`
	MovimientoRealizado bool                                 `mapstructure:"movimientoRealizado" json:"movimientoRealizado"`
	CartaEntregada      bool                                 `mapstructure:"-" json:"cartaEntregada"`
	Cartas              []Mano                               `mapstructure:"-" json:"cartas"`
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
	p.TurnoJugador = 0
	p.Fase = faseRefuerzo
	p.Empezada = true
	p.UltimoTurno = time.Now().UTC().Format(time.RFC3339)
	p.MovimientoRealizado = false
	p.CartaEntregada = false
	// Inicializar las manos de los jugadores
	for i := 0; i < len(p.Jugadores); i++ {
		p.Cartas = append(p.Cartas, Mano{0, 0, 0})
	}
	p.AsignarRefuerzos(0)
	return nil
}

/*
	Restaurar da un valor a los atributos de p que no se guardan en la base de
	datos, y cambia el momento de inicio del último turno por el momento actual
	para evitar que haya jugadores que pierdan su turno por una caída del servidor.
*/
func (p *Partida) Restaurar() {
	p.Conexiones = sync.Map{}
	p.UltimoTurno = time.Now().UTC().Format(time.RFC3339)
	p.Mensajes = make(chan mensajesInternos.MensajePartida, maxMensajes)
}

/*
	AnularInicio anula el inicio de una partida.
*/
func (p *Partida) AnularInicio() {
	p.Empezada = false
	p.Territorios = []Territorio{}
	p.UltimoTurno = ""
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

/*
	ExpulsarDePartida elimina de una partida sin empezar a un jugador que no
	sea el creador de la misma, o marca como muerto a un jugador en una partida
	empezada.
	Devuelve error si hay algún problema.
*/
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
		return errors.New("mal uso de la función: no se puede expulsar de la " +
			"partida al creador")
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
			tropasTerritorio[i][j] = numTropasIniciales / len(tropasTerritorio[i])
			if j < numTropasIniciales%len(tropasTerritorio[i]) {
				tropasTerritorio[i][j]++
			}
		}
	}
	return tropasTerritorio
}

/*
	Refuerzo coloca tropas de un jugador en un territorio y devuelve un json
	con el estado en el que ha quedado el territorio y un campo "_tipoMensaje".
	Si ocurre algún error lo devuelve en formato json.
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
		return mensajes.ErrorJsonPartida("El territorio no te pertenece", 1)
	}
	// Comprobar que tiene suficientes refuerzos
	if p.Jugadores[idJugador].Refuerzos < refuerzos {
		return mensajes.ErrorJsonPartida("Se están intentando asignar más tropas"+
			" que las disponibles", 1)
	}
	if refuerzos < 0 {
		return mensajes.ErrorJsonPartida("No se puede asignar un número negativo"+
			" de tropas a un territorio", 1)
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
	Ataque realiza un ataque entre los territorios indicados y devuelve un json
	con el estado de los territorios involucrados después del ataque y los
	valores obtenidos en los dados.
	Si ocurre algún error lo devuelve en formato json.
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
		return mensajes.ErrorJsonPartida("No puedes atacar desde un territorio"+
			" que no te pertenece", 1)
	}
	// Comprobar que el territorio al que ataca no pertenece al jugador
	if p.Territorios[idDestino].IdJugador == idJugador {
		return mensajes.ErrorJsonPartida("No puedes atacar a un territorio"+
			" que ya te pertenece", 1)
	}
	// Comprobar que los territorios son adyacentes
	if !p.sonAdyacentes(idOrigen, idDestino) {
		return mensajes.ErrorJsonPartida("Los territorios no están conectados", 1)
	}
	// Comprobar que se tienen suficientes tropas (y el origen no queda vacío)
	if p.Territorios[idOrigen].NumTropas <= atacantes {
		return mensajes.ErrorJsonPartida("No tienes tropas suficientes, siempre"+
			" debe quedar al menos una tropa en el territorio de origen", 1)
	}
	// Inicializar el string vacío
	infoRefuerzos := ""
	// Algoritmo ataque
	defensores := min(maxDadosDefensa, p.Territorios[idDestino].NumTropas)
	dadosAtaque := make([]int, min(maxDadosAtaque, atacantes))
	dadosDefensa := make([]int, defensores)
	// Lanzar los dados del atacante
	for i := 0; i < maxDadosAtaque && i < atacantes; i++ {
		dadosAtaque[i] = rand.Intn(6) + 1
	}
	// Lanzar los dados del defensor
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
		if !p.CartaEntregada {
			p.Cartas[idJugador].Robar(rand.Intn(3))
			p.CartaEntregada = true
			infoRefuerzos = p.Cartas[idJugador].Informar()
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
		"dadosDestino":      dadosDefensa,
		"infoRefuerzos":     infoRefuerzos,
	}
}

/*
	existeRuta devuelve si existe un camino que conecte los territorios idOrigen
	e idDestino pasando solo por territorios que pertenecen al jugador idJugador.
	explorados es la lista de territorios por los que ya se ha pasado antes de
	llegar a idOrigen.
*/
func (p *Partida) existeRuta(idOrigen, idDestino, idJugador int, explorados []int) bool {
	// Caso base
	if idOrigen == idDestino {
		return true
	}
	// Si ya se ha explorado el territorio se devuelve falso
	for _, e := range explorados {
		if idOrigen == e {
			return false
		}
	}
	// Añadir el territorio origen a explorados
	explorados = append(explorados, idOrigen)
	// Comprobar todas las conexiones
	for _, i := range infoMapa[idOrigen].Conexiones {
		// Solo considerar territorios del usuario
		if p.Territorios[i].IdJugador == idJugador {
			res := p.existeRuta(i, idDestino, idJugador, explorados)
			if res {
				return true
			}
		}
	}
	return false
}

/*
	sonAdyacentes devuelve si dos territorios están conectados.
*/
func (p *Partida) sonAdyacentes(idOrigen, idDestino int) bool {
	for _, i := range infoMapa[idOrigen].Conexiones {
		if i == idDestino {
			return true
		}
	}
	return false
}

/*
	Movimiento mueve tropas entre los territorios indicados y devuelve un json
	con el estado de los territorios después del movimiento.
	Si ocurre algún error lo devuelve en formato json.
*/
func (p *Partida) Movimiento(idOrigen, idDestino, idJugador, tropas int) mensajes.JsonData {
	// Comprobar que la fase es correcta
	if p.Fase != faseMovimiento {
		return mensajes.ErrorJsonPartida("No estás en la fase de movimiento", 1)
	}
	// Comprobar que se puede realizar el movimiento
	if p.MovimientoRealizado {
		return mensajes.ErrorJsonPartida("Solo puedes realizar un movimiento por "+
			"fase de movimiento", 1)
	}
	// Comprobar que es el turno del jugador
	if p.TurnoJugador != idJugador {
		return mensajes.ErrorJsonPartida("No es tu turno", 1)
	}
	// Comprobar que el territorio de origen pertenece al jugador
	if p.Territorios[idOrigen].IdJugador != idJugador {
		return mensajes.ErrorJsonPartida("No se pueden mover tropas de un"+
			" territorio que no te pertenece", 1)
	}
	// Comprobar que el territorio destino pertenece al jugador
	if p.Territorios[idDestino].IdJugador != idJugador {
		return mensajes.ErrorJsonPartida("No se pueden mover tropas a un"+
			" territorio que no te pertenece", 1)
	}
	// Comprobar que no se muevan al mismo sitio
	if idOrigen == idDestino {
		return mensajes.ErrorJsonPartida("No puedes mover tropas al mismo territorio", 1)
	}
	// Comprobar que sea mayor que 0 el número de tropas
	if idOrigen == idDestino {
		return mensajes.ErrorJsonPartida("No puedes mover 0 tropas", 1)
	}
	// Comprobar que existe ruta entre territorios del jugador
	if !p.existeRuta(idOrigen, idDestino, idJugador, []int{}) {
		return mensajes.ErrorJsonPartida("No existe ruta entre territorios", 1)
	}
	// Comprobar que se tienen suficientes tropas (y el origen no queda vacío)
	if p.Territorios[idOrigen].NumTropas <= tropas {
		return mensajes.ErrorJsonPartida("No tienes tropas suficientes, siempre"+
			" debe quedar al menos una tropa en el territorio de origen", 1)
	}
	// Guardar que en esta fase se ha realizado un movmiento
	p.MovimientoRealizado = true
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

/*
	AsignarRefuerzos calcula la cantidad de refuerzos que le corresponden al
	jugador indicado según el estado de la partida y se los asigna.
*/
func (p *Partida) AsignarRefuerzos(id int) {
	// Dar el número de refuerzos que corresponde por número de territorios
	p.Jugadores[id].Refuerzos = p.Jugadores[id].NumTerritorios / 3
	if p.Jugadores[id].Refuerzos < 3 {
		p.Jugadores[id].Refuerzos = 3
	}
	// Dar el número de refuerzos que corresponde por continentes
	cuenta := [numContinentes]int{0, 0, 0, 0, 0, 0}
	for i := range p.Territorios {
		if p.Territorios[i].IdJugador == id {
			cuenta[infoMapa[i].Continente]++
		}
	}
	for i := range cuenta {
		if cuenta[i] == paises[i] {
			p.Jugadores[id].Refuerzos += bonos[i]
		}
	}
	// Dar refuerzos correspondientes a grupos de cartas
	p.Jugadores[id].Refuerzos += p.Cartas[id].Negociar()
}

/*
	ObtenerPosicionJugador devuelve el identificador dentro de la partida del
	jugador indicado.
	Si no está en la partida, devuelve -1.
*/
func (p *Partida) ObtenerPosicionJugador(id int) int {
	for i := range p.Jugadores {
		if p.Jugadores[i].Id == id {
			return i
		}
	}
	return -1
}

/*
	PasarTurno avanza de turno la partida y asigna los refuerzos que le
	corresponden al siguiente jugador.
*/
func (p *Partida) PasarTurno() mensajes.JsonData {
	var res mensajes.JsonData

	p.Fase = faseRefuerzo
	p.TurnoActual++
	// Calcular el jugador de cual es el turno
	p.TurnoJugador = (p.TurnoJugador + 1) % len(p.Jugadores)
	for !p.Jugadores[p.TurnoJugador].SigueVivo {
		p.TurnoJugador = (p.TurnoJugador + 1) % len(p.Jugadores)
	}
	// Calcular el nuevo valor de los refuerzos para el jugador al que le toca
	p.AsignarRefuerzos(p.TurnoJugador)
	// Nueva marca temporal del ultimo turno
	p.UltimoTurno = time.Now().UTC().Format(time.RFC3339)
	// Codificar los datos de la partida en formato json
	mapstructure.Decode(p, &res)
	res["_tipoMensaje"] = "p"
	return res
}

/*
	AvanzarFase avanza la fase en la que se encuentra el jugador y devuelve un
	json de confirmación.
	Si ocurre algún error lo devuelve en formato json.
*/
func (p *Partida) AvanzarFase(jugador int) mensajes.JsonData {

	if p.TurnoJugador != jugador {
		return mensajes.ErrorJsonPartida("No es tu turno", 1)
	}

	res := mensajes.JsonData{"_tipoMensaje": "f"}

	switch p.Fase {
	case faseRefuerzo:
		if p.Jugadores[jugador].Refuerzos > 0 {
			return mensajes.ErrorJsonPartida("Aún te quedan "+
				strconv.Itoa(p.Jugadores[jugador].Refuerzos)+
				" refuerzos por colocar", 1)
		}
		p.Fase++
		p.CartaEntregada = false
		return res
	case faseAtaque:
		p.Fase++
		p.MovimientoRealizado = false
		return res
	case faseMovimiento:
		return p.PasarTurno()
	}

	return mensajes.ErrorJsonPartida("La partida no está empezada", 1)
}

/*
	JugadoresRestantes devuelve el número de jugadores que quedan en la partida.
*/
func (p *Partida) JugadoresRestantes() int {
	respuesta := 0
	for _, j := range p.Jugadores {
		if j.SigueVivo {
			respuesta++
		}
	}
	return respuesta
}

/*
	FinalizarPartida devuelve el mensaje de fin y el identificador
	de usuario del ganador, si se puede acabar la partida.
	En caso contrario, devuelve el error ocurrido.
*/
func (p *Partida) FinalizarPartida() (mensajes.JsonData, int, error) {
	respuesta := mensajes.JsonData{}
	ganador := p.Jugadores[p.TurnoJugador].Id
	if p.JugadoresRestantes() > 1 {
		return respuesta, ganador, errors.New("queda más de un jugador con " +
			"territorios, no se puede terminar la partida")
	}
	respuesta = mensajes.JsonData{
		"_tipoMensaje": "t",
		"ganador":      p.TurnoJugador,
		"riskos":       50,
	}
	return respuesta, ganador, nil
}

// Funciones de envío de correos

/*
	EnviarCorreoTurno envía un correo al jugador al que le corresponde el turno
	en la partida `p` indicando el nombre de la partida.
	Solo envía el correo si el usuario ha indicado que quiere recibir las
	notificaciones por correo y no está conectado a la partida desde ningún
	dispositivo.
	Para enviar el correo utiliza el servidor, dirección y contraseña indicados
	como parámetros, y si no consigue enviarlo devuelve el error ocurrido.
*/
func (p *Partida) EnviarCorreoTurno(smtpServer, smtpPort, correo, clave string) error {
	if p.Jugadores[p.TurnoJugador].Correo == "" {
		return errors.New("este usuario no recibe correos")
	}

	if _, ok := p.Conexiones.Load(p.Jugadores[p.TurnoJugador].Id); ok {
		return errors.New("el usuario está conectado")
	}

	e := email.NewEmail()
	e.From = "PixelRisk <" + correo + ">"
	e.To = []string{p.Jugadores[p.TurnoJugador].Correo}
	e.Subject = "Notificación de turno"
	e.Text = []byte("Es tu turno en la partida " + p.Nombre + "!")
	err := e.Send(smtpServer+":"+smtpPort, smtp.PlainAuth("", correo, clave, smtpServer))
	if err != nil {
		return err
	}

	return nil
}

/*
	EnviarCorreoFinPartida envía un correo a todos los jugadores de la partida
	`p` que no estén conectados a ella desde ningún dispositivo y hayan indicado
	que quieren recibir las notificaciones por correo.
	Para preservar la privacidad de los usuarios, los destinatarios están
	ocultos y el resto no pueden ver su dirección de correo electrónico.
	Este correo contiene el nombre de la partida y el del ganador.
	Para enviar el correo utiliza el servidor, dirección y contraseña indicados
	como parámetros, y si no consigue enviarlo devuelve el error ocurrido.
*/
func (p *Partida) EnviarCorreoFinPartida(smtpServer, smtpPort, correo, clave string) error {
	destinatariosId := []int{}
	destinatariosCorreo := []string{}

	for i, j := range p.Jugadores {
		if j.Correo != "" {
			destinatariosId = append(destinatariosId, i)
		}
	}

	if len(destinatariosId) == 0 {
		return errors.New("ningún jugador de la partida recibe correos")
	}

	for _, i := range destinatariosId {
		if _, ok := p.Conexiones.Load(p.Jugadores[i].Id); !ok {
			destinatariosCorreo = append(destinatariosCorreo, p.Jugadores[i].Correo)
		}
	}
	if len(destinatariosCorreo) == 0 {
		return errors.New("todos los usuarios están conectados")
	}

	e := email.NewEmail()
	e.From = "PixelRisk <" + correo + ">"
	e.Subject = "Fin de partida"
	e.Text = []byte("¡" + p.Jugadores[p.TurnoJugador].Nombre +
		" ha ganado la partida \"" + p.Nombre + "\"!")
	e.Bcc = destinatariosCorreo
	err := e.Send(smtpServer+":"+smtpPort, smtp.PlainAuth("", correo, clave, smtpServer))
	if err != nil {
		return err
	}

	return nil
}

// Funciones de envío de mensaje a través de WebSockets

/*
	EnviarATodos envía el mensaje pasado como parámetro a todos los usuarios
	conectados a la partida `p`, por el websocket correspondiente.
*/
func (p *Partida) EnviarATodos(mensaje mensajes.JsonData) {
	for _, jugador := range p.Jugadores {
		ws, ok := p.Conexiones.Load(jugador.Id)
		if ok {
			ws.(*websocket.Conn).WriteJSON(mensaje)
		}
	}
}

/*
	EnviarError envía un mensaje de error al usuario con el identificador global
	indicado, por el websocket correspondiente en la partida `p`.
*/
func (p *Partida) EnviarError(idUsuario, code int, err string) {
	ws, ok := p.Conexiones.Load(idUsuario)
	if ok {
		ws.(*websocket.Conn).WriteJSON(mensajes.ErrorJsonPartida(err, code))
	}
}

/*
	Enviar envía el mensaje pasado como parámetro al usuario con el identificador
	global indicado, por el websocket correspondiente en la partida `p`.
*/
func (p *Partida) Enviar(idUsuario int, mensaje mensajes.JsonData) {
	ws, ok := p.Conexiones.Load(idUsuario)
	if ok {
		ws.(*websocket.Conn).WriteJSON(mensaje)
	}
}

// FUNCIONES AUXILIARES PARA EL MANEJO DE ARRAYS

/*
	borrar elimina de un slice de jugadores el primero que encuentra cuyo
	identificador coincide con `valor` y devuelve el nuevo slice.
	Si dicho jugador no está en la lista, se devuelve sin modificar.
*/
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
