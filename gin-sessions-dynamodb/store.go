package sessions

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb"
	"github.com/nguyengg/go-aws-commons/ddb/model"
)

type store struct {
	name    string
	manager *ddb.Manager
	m       *model.Model
	t       reflect.Type
	v       interface{}
}

func (s *store) get(c *gin.Context) interface{} {
	if s.v != nil {
		return s.v
	}

	sid, err := c.Cookie(s.name)
	if errors.Is(err, http.ErrNoCookie) {
		return s.new()
	}

	v := reflect.New(s.t)
	s.v = v.Interface()

	if v, err = s.m.HashKey.Get(v.Elem()); err != nil {
		panic(err)
	} else {
		v.SetString(sid)
	}

	if _, err = s.manager.Get(c, s.v, s.v); err != nil {
		_ = c.AbortWithError(http.StatusBadGateway, err)
		return nil
	}

	return s.v
}

func (s *store) new() interface{} {
	v := reflect.New(s.t)
	s.v = v.Interface()

	if v, err := s.m.HashKey.Get(v.Elem()); err != nil {
		panic(err)
	} else {
		v.SetString(NewURLSafeId())
	}

	return s.v
}

func (s *store) save(c *gin.Context, opts *SaveOptions) {
	// if the session value is nil then that means there has not been any changes to the session, so skip the save.
	if s.v == nil {
		return
	}

	if _, err := s.manager.Put(c, s.v); err != nil {
		_ = c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	s.v = reflect.New(s.t)
	v, err := s.m.HashKey.Get(reflect.ValueOf(s.v))
	if err != nil {
		panic(err)
	}

	c.SetSameSite(opts.SameSite)
	c.SetCookie(s.name, v.String(), opts.MaxAge, opts.Path, opts.Domain, opts.Secure, opts.HttpOnly)
}
