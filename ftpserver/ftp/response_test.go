package ftp

import (
	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ResponseSuite struct{}

func (s *ResponseSuite) TestMarshalText(t sweet.T) {
	r := NewResponse(200, "Hello there!")
	d, err := r.MarshalText()
	Expect(err).To(BeNil())
	Expect(d).To(Equal([]byte("200 Hello there!")))
}
