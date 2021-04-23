package baseDatos

const (
	europa       = 0
	africa       = 1
	asia         = 2
	oceania      = 3
	americaSur   = 4
	americaNorte = 5
)

const numContinentes = 6
const numTerritorios = 42

var paises = [numContinentes]int{7, 6, 12, 4, 4, 9}
var bonos = [numContinentes]int{5, 3, 7, 2, 2, 5}

type infoTerritorio struct {
	Conexiones []int
	Continente int
}

var infoMapa = [numTerritorios]infoTerritorio{
	{Conexiones: []int{1, 2, 4, 7}, Continente: europa},          //id 0
	{Conexiones: []int{0, 2, 3, 4, 6}, Continente: europa},       // id 1
	{Conexiones: []int{0, 1, 6, 7, 8, 13}, Continente: europa},   // id 2
	{Conexiones: []int{1, 4, 6, 5}, Continente: europa},          // id 3
	{Conexiones: []int{0, 1, 3, 5}, Continente: europa},          // id 4
	{Conexiones: []int{3, 4, 41}, Continente: europa},            // id 5
	{Conexiones: []int{1, 2, 3, 13, 15, 16}, Continente: europa}, // id 6

	{Conexiones: []int{0, 2, 8, 9, 10, 30}, Continente: africa},   // id 7
	{Conexiones: []int{2, 7, 9, 13}, Continente: africa},          // id 8
	{Conexiones: []int{7, 8, 10, 11, 12, 13}, Continente: africa}, // id 9
	{Conexiones: []int{7, 9, 11}, Continente: africa},             // id 10
	{Conexiones: []int{9, 10, 12}, Continente: africa},            // id 11
	{Conexiones: []int{9, 11}, Continente: africa},                // id 12

	{Conexiones: []int{2, 6, 8, 9, 14, 15}, Continente: asia},     // id 13
	{Conexiones: []int{13, 15, 22, 24}, Continente: asia},         // id 14
	{Conexiones: []int{6, 13, 14, 16, 22}, Continente: asia},      // id 15
	{Conexiones: []int{6, 15, 17, 22}, Continente: asia},          // id 16
	{Conexiones: []int{16, 18, 20, 21, 22}, Continente: asia},     // id 17
	{Conexiones: []int{17, 19, 20}, Continente: asia},             // id 18
	{Conexiones: []int{18, 20, 21, 23, 40}, Continente: asia},     // id 19
	{Conexiones: []int{17, 18, 19, 21}, Continente: asia},         // id 20
	{Conexiones: []int{17, 19, 20, 22, 23}, Continente: asia},     // id 21
	{Conexiones: []int{14, 15, 16, 17, 21, 24}, Continente: asia}, // id 22
	{Conexiones: []int{19, 21}, Continente: asia},                 // id 23
	{Conexiones: []int{14, 22, 25}, Continente: asia},             // id 24

	{Conexiones: []int{24, 26, 28}, Continente: oceania}, // id 25
	{Conexiones: []int{25, 27, 28}, Continente: oceania}, // id 26
	{Conexiones: []int{26, 28}, Continente: oceania},     // id 27
	{Conexiones: []int{25, 26, 27}, Continente: oceania}, // id 28

	{Conexiones: []int{30, 31}, Continente: americaSur},        // id 29
	{Conexiones: []int{7, 29, 31, 32}, Continente: americaSur}, // id 30
	{Conexiones: []int{29, 30, 32}, Continente: americaSur},    // id 31
	{Conexiones: []int{30, 31, 33}, Continente: americaSur},    // id 32

	{Conexiones: []int{32, 34, 35}, Continente: americaNorte},         // id 33
	{Conexiones: []int{33, 35, 36, 37}, Continente: americaNorte},     // id 34
	{Conexiones: []int{33, 34, 37, 38}, Continente: americaNorte},     // id 35
	{Conexiones: []int{34, 37, 41}, Continente: americaNorte},         // id 36
	{Conexiones: []int{34, 35, 36, 38, 39}, Continente: americaNorte}, // id 37
	{Conexiones: []int{35, 37, 39, 40}, Continente: americaNorte},     // id 38
	{Conexiones: []int{37, 38, 40, 41}, Continente: americaNorte},     // id 39
	{Conexiones: []int{19, 38, 39}, Continente: americaNorte},         // id 40
	{Conexiones: []int{5, 36, 39}, Continente: americaNorte},          // id 41
}
