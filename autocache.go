package autocache

import "sync"

type LookupFuncType func(string) (string, error)

type Cache struct {
	head       *node
	tail       *node
	hash       map[string]*node
	lookupFunc LookupFuncType
	count      int
	maxSize    int
	mutex      sync.Mutex
}

type node struct {
	key  string
	val  string
	next *node
	prev *node
}

func New(maxSize int, lookupFunc LookupFuncType) *Cache {
	c := &Cache{}
	c.maxSize = maxSize
	c.hash = make(map[string]*node)
	c.lookupFunc = lookupFunc
	return c
}

func (c *Cache) Get(key string) (string, error) {

	valNode, ok := c.hash[key]
	chopTail := false
	if !ok {
		val, err := c.lookupFunc(key)
		if err != nil {
			return val, err
		}
		valNode = &node{key, val, nil, nil}
		c.hash[key] = valNode
		if c.count >= c.maxSize {
			chopTail = true
		} else {
			c.count++
		}
	}

	//this lock may be overzealous
	//erring on the side of caution for now
	c.mutex.Lock()

	//remove valnode from current place in tree
	if valNode.prev != nil {
		valNode.prev.next = valNode.next
	}

	if valNode.next != nil {
		valNode.next.prev = valNode.prev
	}

	//if valnode was tail, clean up
	if c.tail == valNode {
		c.tail = valNode.prev
		if c.tail != nil {
			c.tail.next = nil
		}
	}

	//set valnode as head
	valNode.prev = nil
	if c.head != nil {
		c.head.prev = valNode
	}
	valNode.next = c.head
	c.head = valNode

	//on first get, set as tail too
	if c.tail == nil {
		c.tail = valNode
	}

	if chopTail {
		delete(c.hash, c.tail.key)
		c.tail = c.tail.prev
		c.tail.next = nil
	}

	c.mutex.Unlock()
	return valNode.val, nil
}
