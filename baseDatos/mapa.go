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
	{Conexiones: []int{1, 2, 4, 7}, Continente: europa},
	{Conexiones: []int{0, 2, 3, 4, 6}, Continente: europa},
	{Conexiones: []int{0, 1, 6, 7, 8, 13}, Continente: europa},
	{Conexiones: []int{1, 4, 6, 5}, Continente: europa},
	{Conexiones: []int{0, 1, 3, 5}, Continente: europa},
	{Conexiones: []int{3, 4, 41}, Continente: europa},
	{Conexiones: []int{1, 2, 3, 13, 15, 16}, Continente: europa},

	{Conexiones: []int{0, 2, 8, 9, 10, 30}, Continente: africa},
	{Conexiones: []int{2, 7, 9, 13}, Continente: africa},
	{Conexiones: []int{7, 8, 10, 11, 12, 13}, Continente: africa},
	{Conexiones: []int{7, 9, 11}, Continente: africa},
	{Conexiones: []int{9, 10, 12}, Continente: africa},
	{Conexiones: []int{9, 11}, Continente: africa},

	{Conexiones: []int{2, 6, 8, 9, 14, 15}, Continente: asia},
	{Conexiones: []int{13, 15, 22, 24}, Continente: asia},
	{Conexiones: []int{6, 13, 14, 16, 22}, Continente: asia},
	{Conexiones: []int{6, 15, 17, 22}, Continente: asia},
	{Conexiones: []int{16, 18, 20, 21, 22}, Continente: asia},
	{Conexiones: []int{17, 19, 20}, Continente: asia},
	{Conexiones: []int{18, 20, 21, 23, 40}, Continente: asia},
	{Conexiones: []int{17, 18, 19, 21}, Continente: asia},
	{Conexiones: []int{17, 19, 20, 22, 23}, Continente: asia},
	{Conexiones: []int{14, 15, 16, 17, 21, 24}, Continente: asia},
	{Conexiones: []int{19, 21}, Continente: asia},
	{Conexiones: []int{14, 22, 25}, Continente: asia},

	{Conexiones: []int{24, 26}, Continente: oceania},
	{Conexiones: []int{25, 27, 28}, Continente: oceania},
	{Conexiones: []int{26, 28}, Continente: oceania},
	{Conexiones: []int{26, 27}, Continente: oceania},

	{Conexiones: []int{30, 31}, Continente: americaSur},
	{Conexiones: []int{7, 29, 31, 32}, Continente: americaSur},
	{Conexiones: []int{29, 30, 32}, Continente: americaSur},
	{Conexiones: []int{30, 31, 33}, Continente: americaSur},

	{Conexiones: []int{32, 34, 35}, Continente: americaNorte},
	{Conexiones: []int{33, 35, 36, 37}, Continente: americaNorte},
	{Conexiones: []int{33, 34, 37, 38}, Continente: americaNorte},
	{Conexiones: []int{34, 37, 41}, Continente: americaNorte},
	{Conexiones: []int{34, 35, 36, 38, 39}, Continente: americaNorte},
	{Conexiones: []int{35, 37, 39, 40}, Continente: americaNorte},
	{Conexiones: []int{37, 38, 40, 41}, Continente: americaNorte},
	{Conexiones: []int{19, 38, 39}, Continente: americaNorte},
	{Conexiones: []int{5, 36, 39}, Continente: americaNorte},
}
