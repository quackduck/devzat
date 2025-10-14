package main

type AdminLogWriter struct{}

func (alw AdminLogWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
