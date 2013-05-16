package tree

import ()

// An easy to build priority queue
type prioq struct {
	prios               [][]*OrderNode
	lowPrice, highPrice int64
}

func mkPrioq(lowPrice, highPrice int64) *prioq {
	prios := make([][]*OrderNode, highPrice-lowPrice+1)
	return &prioq{prios: prios, lowPrice: lowPrice, highPrice: highPrice}
}

func (q *prioq) push(o *OrderNode) {
	idx := o.Price() - q.lowPrice
	prio := q.prios[idx]
	prio = append(prio, o)
	q.prios[idx] = prio
}

func (q *prioq) peekMax() *OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := len(q.prios) - 1; i >= 0; i-- {
		switch {
		case len(q.prios[i]) > 0:
			return q.prios[i][0]
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMax() *OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := len(q.prios) - 1; i >= 0; i-- {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) peekMin() *OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := 0; i < len(q.prios); i++ {
		switch {
		case len(q.prios[i]) > 0:
			return q.prios[i][0]
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMin() *OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := 0; i < len(q.prios); i++ {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) pop(i int) *OrderNode {
	prio := q.prios[i]
	o := prio[0]
	prio = prio[1:]
	q.prios[i] = prio
	return o
}

func (q *prioq) cancel(guid int64) *OrderNode {
	for i := range q.prios {
		priceQ := q.prios[i]
		for j := range priceQ {
			o := priceQ[j]
			if o.Guid() == guid {
				priceQ = append(priceQ[0:j], priceQ[j+1:]...)
				q.prios[i] = priceQ
				return o
			}
		}
	}
	return nil
}
