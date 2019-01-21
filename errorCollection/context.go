package errorCollection

import "sync"

type Context struct {
	M   *sync.RWMutex
	Map map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		M:   &sync.RWMutex{},
		Map: make(map[string]interface{}, 0),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.M.Lock()
	defer c.M.Unlock()
	c.Map[key] = value
}

func (c *Context) Get(key string) interface{} {
	c.M.RLock()
	defer c.M.RUnlock()
	return c.Map[key]
}
func (c *Context) GetString(key string) string {
	c.M.RLock()
	defer c.M.RUnlock()
	return c.Map[key].(string)
}
func (c *Context) GetBool(key string) bool {
	c.M.RLock()
	defer c.M.RUnlock()
	return c.Map[key].(bool)
}
func (c *Context) GetInt(key string) int {
	c.M.RLock()
	defer c.M.RUnlock()
	return c.Map[key].(int)
}
