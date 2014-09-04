// Commands from http://redis.io/commands#generic

package miniredis

import (
	"strconv"

	"github.com/bsm/redeo"
)

// Expire value. As set by the client. 0 if not set.
func (m *Miniredis) Expire(k string) int {
	m.Lock()
	defer m.Unlock()
	return m.expire[k]
}

// SetExpire sets expiration of a key.
func (m *Miniredis) SetExpire(k string, ex int) {
	m.Lock()
	defer m.Unlock()
	m.expire[k] = ex
}

// commandsGeneric handles EXPIRE, TTL, PERSIST
func commandsGeneric(m *Miniredis, srv *redeo.Server) {
	srv.HandleFunc("EXPIRE", func(out *redeo.Responder, r *redeo.Request) error {
		if len(r.Args) != 2 {
			out.WriteErrorString("usage error")
			return nil
		}
		key := r.Args[0]
		value := r.Args[1]
		i, err := strconv.Atoi(value)
		if err != nil {
			out.WriteErrorString("value error")
			return nil
		}
		m.Lock()
		defer m.Unlock()
		// Key must be present.
		if _, ok := m.keys[key]; !ok {
			out.WriteZero()
			return nil
		}
		m.expire[key] = i
		out.WriteOne()
		return nil
	})

	srv.HandleFunc("TTL", func(out *redeo.Responder, r *redeo.Request) error {
		key := r.Args[0]
		m.Lock()
		defer m.Unlock()
		if _, ok := m.keys[key]; !ok {
			// No such key
			out.WriteInt(-2)
			return nil
		}

		value, ok := m.expire[key]
		if !ok {
			// No expire value
			out.WriteInt(-1)
			return nil
		}
		out.WriteInt(value)
		return nil
	})

	srv.HandleFunc("PERSIST", func(out *redeo.Responder, r *redeo.Request) error {
		key := r.Args[0]
		m.Lock()
		defer m.Unlock()
		if _, ok := m.keys[key]; !ok {
			// No such key
			out.WriteInt(0)
			return nil
		}

		_, ok := m.expire[key]
		if !ok {
			// No expire value
			out.WriteInt(0)
			return nil
		}
		delete(m.expire, key)
		out.WriteInt(1)
		return nil
	})

	// MULTI is a no-op
	srv.HandleFunc("MULTI", func(out *redeo.Responder, r *redeo.Request) error {
		out.WriteOK()
		return nil
	})

	// EXEC is a no-op
	srv.HandleFunc("EXEC", func(out *redeo.Responder, r *redeo.Request) error {
		out.WriteNil()
		return nil
	})

	srv.HandleFunc("DEL", func(out *redeo.Responder, r *redeo.Request) error {
		m.Lock()
		defer m.Unlock()

		count := 0
		for _, key := range r.Args {
			if _, ok := m.keys[key]; !ok {
				continue
			}
			delete(m.keys, key)
			delete(m.expire, key)
			// These are not strictly needed:
			delete(m.stringKeys, key)
			delete(m.hashKeys, key)
			count++
		}
		out.WriteInt(count)
		return nil
	})
}
