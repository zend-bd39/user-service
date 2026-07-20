package interceptor

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecoverInterceptorTest struct {
	suite.Suite
}

func (s *RecoverInterceptorTest) TestRecovery() {
	t := s.T()
	t.Run("success -recovery from panic", func(t *testing.T) {
		recovery := RecoveryInterceptor()
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/Register",
		}
		handler := func(ctx context.Context, req any) (any, error) {
			panic(fmt.Sprintf("Test panic untuk recovery %s", "PANIC"))
		}
		resp, err := recovery(context.Background(), "dummy", info, handler)
		assert.Error(t, err)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.Internal)
		assert.Nil(t, resp)
	})
}
func TestRecoverInterceptorTest(t *testing.T) {
	suite.Run(t, new(RecoverInterceptorTest))
}