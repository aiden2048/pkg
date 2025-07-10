package runtime

import "sync"

func NewWithUsageMu(internalMu sync.Locker, onUsageAdd func(), onUsageSub func()) *WithUsageMu {
	return &WithUsageMu{
		internalMu: internalMu,
		onUsageAdd: onUsageAdd,
		onUsageSub: onUsageSub,
	}
}

type WithUsageMu struct {
	outerMu    sync.RWMutex
	internalMu sync.Locker
	UsedNum    uint32
	onUsageSub func()
	onUsageAdd func()
}

func (m *WithUsageMu) Lock() {
	m.internalMu.Lock()
	m.UsedNum++
	m.onUsageAdd()
	m.internalMu.Unlock()

	m.outerMu.Lock()
}

func (m *WithUsageMu) Unlock() {
	m.internalMu.Lock()
	m.UsedNum--
	m.onUsageSub()
	m.internalMu.Unlock()

	m.outerMu.Unlock()
}

func (m *WithUsageMu) RLock() {
	m.internalMu.Lock()
	m.UsedNum++
	m.onUsageAdd()
	m.internalMu.Unlock()

	m.outerMu.RLock()
}

func (m *WithUsageMu) RUnlock() {
	m.internalMu.Lock()
	m.UsedNum--
	m.onUsageSub()
	m.internalMu.Unlock()

	m.outerMu.RUnlock()
}

type MulElemMuFactory struct {
	elemMuMap map[interface{}]*WithUsageMu
	opMapMu   sync.RWMutex
}

func NewMulElemMuFactory() *MulElemMuFactory {
	return &MulElemMuFactory{
		elemMuMap: map[interface{}]*WithUsageMu{},
	}
}

func (m *MulElemMuFactory) MakeOrGetSpecElemMu(elem interface{}) *WithUsageMu {
	m.opMapMu.RLock()
	mu, ok := m.elemMuMap[elem]
	m.opMapMu.RUnlock()
	if ok {
		return mu
	}

	m.opMapMu.Lock()
	defer m.opMapMu.Unlock()

	mu, ok = m.elemMuMap[elem]
	if !ok {
		mu = NewWithUsageMu(
			&m.opMapMu,
			func() {
				if _, ok := m.elemMuMap[elem]; ok {
					return
				}
				m.elemMuMap[elem] = mu
			},
			func() {
				if mu.UsedNum > 0 {
					return
				}
				delete(m.elemMuMap, elem)
			},
		)

		m.elemMuMap[elem] = mu
	}
	return mu
}
