package claudeagent

import (
	"context"
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// exitedTransport accepts writes but produces no output, like a CLI that
// exited during startup.
type exitedTransport struct{ *mockTransport }

func (exitedTransport) ReadMessages(context.Context) iter.Seq2[Message, error] {
	return func(func(Message, error) bool) {}
}

// TestConnectFailsWhenOutputEndsBeforeInitializeReply checks that Connect
// returns once the CLI's output ends instead of waiting out its context.
func TestConnectFailsWhenOutputEndsBeforeInitializeReply(t *testing.T) {
	client, err := NewClient(WithTransport(exitedTransport{newMockTransport(0)}))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	err = client.Connect(ctx)

	var failed *ErrSubprocessFailed
	require.ErrorAs(t, err, &failed)
	require.Less(t, time.Since(start), 2*time.Second)
}

// TestWaitForSDKResponsePrefersReplyRoutedBeforeOutputEnded checks that a
// reply already routed is returned even though the output has since ended.
func TestWaitForSDKResponsePrefersReplyRoutedBeforeOutputEnded(t *testing.T) {
	p := NewProtocol(newMockTransport(0), &Options{})
	ch := make(chan SDKControlResponse, 1)
	p.pendingReqs.Store("req_1", ch)
	ch <- SDKControlResponse{Type: "control_response"}
	p.endOutput()

	resp, err := p.waitForSDKResponse(context.Background(), "req_1", ch)
	require.NoError(t, err)
	require.Equal(t, "control_response", resp.Type)
}
