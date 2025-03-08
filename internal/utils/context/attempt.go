package context

import "context"

var attemptKey = "attempt"

func SetAttempt(ctx context.Context, attempt int) context.Context {
	return context.WithValue(ctx, attemptKey, attempt)
}

func GetAttempt(ctx context.Context) int {
	attempt, _ := ctx.Value(attemptKey).(int)

	return attempt
}
