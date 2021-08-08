package api

type MatcherMap struct {
	keys   map[uint64]*Matcher    // index -> Key
	values map[uint64]interface{} // index -> Value
}

func NewMatcherMap() *MatcherMap {
	return &MatcherMap{
		keys:   map[uint64]*Matcher{},
		values: map[uint64]interface{}{},
	}
}

func (m *MatcherMap) Get(key *Matcher) interface{} {
	return m.values[key.MapIndex()]
}

func (m *MatcherMap) Exists(key *Matcher) bool {
	_, ok := m.keys[key.MapIndex()]
	return ok
}

func (m *MatcherMap) Set(key *Matcher, value interface{}) {
	idx := key.MapIndex()
	m.keys[idx] = key
	m.values[idx] = value
}

func (m *MatcherMap) Delete(key *Matcher) {
	idx := key.MapIndex()
	delete(m.keys, idx)
	delete(m.values, idx)
}
