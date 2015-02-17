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

	//if cache does not need to be modified
	if valNode == c.head {
		return valNode.val, nil
	}

	//this lock may not be in the correct or optimal place
	//it seems to work well enough for now.
	c.mutex.Lock()
	//fmt.Printf("\ngetting word '%s'\n%s\n", key, c.debugFullCache())

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

	//fmt.Printf("replacing head %#v\n", c.head)
	//set valnode as head
	valNode.prev = nil
	if c.head != nil {
		c.head.prev = valNode
	}
	valNode.next = c.head
	c.head = valNode
	//fmt.Printf("new head is %#v\n", c.head)

	//on first get, set as tail too
	if c.tail == nil {
		c.tail = valNode
	}

	if chopTail {
		//fmt.Printf("chopping tail %#v\n", c.tail)
		delete(c.hash, c.tail.key)
		c.tail = c.tail.prev
		c.tail.next = nil
		//fmt.Printf("new tail is %#v\n", c.tail)
	}

	c.mutex.Unlock()
	return valNode.val, nil
}

func (n *node) String() string {
	if n == nil {
		return "NIL"
	}

	return "('" + n.key + "','" + n.val + "')"
}

func (n *node) GoString() string {
	if n == nil {
		return "NIL"
	}

	return n.prev.String() + " <- " + n.String() + " -> " + n.next.String()
}

func (c *Cache) debugFullCache() string {
	headStr := ""
	n := c.head
	for {
		if headStr != "" {
			headStr += " -> "
		}
		headStr += n.String()

		if n == nil {
			break
		}
		n = n.next
	}

	tailStr := ""
	n = c.tail
	for {
		if tailStr != "" {
			tailStr += " -> "
		}
		tailStr += n.String()

		if n == nil {
			break
		}
		n = n.prev
	}
	return "FROM HEAD: " + headStr + "\nFROM TAIL: " + tailStr
}
