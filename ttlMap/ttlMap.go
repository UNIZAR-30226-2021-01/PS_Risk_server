package ttlMap

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

/*
	tokenRecuperacion guarda el identificador del usuario asociado y el tiempo
	en minutos que falta para que el token expire.
*/
type tokenRecuperacion struct {
	ttl int
	id  int
}

/*
	TTLmap guarda los tokens de recuperación generados que aún son válidos y el
	tiempo de vida de un nuevo token, y proporciona acceso thread-safe a los tokens.
	Solo permite tener un token asociado a un usuario cada vez.
*/
type TTLmap struct {
	m   map[string]tokenRecuperacion
	l   sync.Mutex
	ttl int
}

/*
	CrearTTLmap crea un nuevo TTLmap con el tiempo de vida por defecto indicado
	en minutos (ttl).
	Si un token supera este tiempo, lo elimina del map.
*/
func CrearTTLmap(ttl int) *TTLmap {
	m := TTLmap{
		m:   make(map[string]tokenRecuperacion),
		l:   sync.Mutex{},
		ttl: ttl,
	}

	go func() {
		for range time.Tick(time.Duration(1) * time.Minute) {
			m.l.Lock()
			for k, v := range m.m {
				v.ttl--
				if v.ttl == 0 {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()

	return &m
}

/*
	NuevoToken genera y guarda un nuevo token de recuperación en m asociado al
	usuario id.
	Devuelve este token, o un error si ha ocurrido alguno.
*/
func (m *TTLmap) NuevoToken(id int) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("no se ha podido generar el token")
	}
	token := hex.EncodeToString(b)

	m.l.Lock()
	for _, v := range m.m {
		if v.id == id {
			m.l.Unlock()
			return "", errors.New("ya se ha solicitado restablecer clave")
		}
	}
	m.m[token] = tokenRecuperacion{ttl: m.ttl, id: id}
	m.l.Unlock()

	return token, nil
}

/*
	ConsumirToken elimina de m el token indicado y devuelve el identificador del
	usuario asociado a él.
	Devuelve error si no existe el token.
*/
func (m *TTLmap) ConsumirToken(t string) (int, error) {
	m.l.Lock()
	if v, ok := m.m[t]; ok {
		delete(m.m, t)
		m.l.Unlock()
		return v.id, nil
	}
	m.l.Unlock()
	return -1, errors.New("token inválido")
}
