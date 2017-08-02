# urlog

```golang
import "github.com/segmentio/ur-log"
```

## Usage

#### func  BuildErrorPacket

```go
func BuildErrorPacket(err error) *raven.Packet
```
BuildErrorPacket builds a sentry packet based on err. Works best if err
implements log.Fielder and stackTracer interfaces

#### func  GetContextualLogger

```go
func GetContextualLogger(ctx context.Context, logger log.Interface, fielder log.Fielder) (context.Context, log.Interface)
```
GetContextualLogger returns updated versions of ctx and logger. If fielder is
not nil, fields from it are merged into fields already stored in ctx. If logger
is nil, a default value of log.Log is assumed. Updated logger contains its old
fields merged with fields from fielder.

#### func  WrapError

```go
func WrapError(ctx context.Context, err error, prefixMessage string) error
```
WrapError returns a wrapped copy of err if it doesn't have a stacktrace, lacks
some fields available in ctx or needs to be prefixed with prefixMessage. If
neither of those is true, original err is returned.
