package main

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_runCmd(t *testing.T) {
	t.Run("returns string echoed to stdout", func(t *testing.T) {
		testString := "hello world"
		want := testString + "\n"

		got, err := runCmd(nil, "echo", testString) //lint:ignore SA1012 nil context

		assert.Equal(t, got, want, "captured string matches input to echo")
		assert.Nil(t, err)
	})

	t.Run("handles context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := runCmd(ctx, "sleep", "10s")

		assert.EqualError(t, ctx.Err(), context.DeadlineExceeded.Error(), "context timed out")
		assert.EqualError(t, err, "signal: killed", "process was killed")
	})

	t.Run("handles cancellation of context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		errc := make(chan error)
		go func() {
			_, err := runCmd(ctx, "sleep", "10s")
			errc <- err
		}()
		cancel()
		err := <-errc

		assert.EqualError(t, ctx.Err(), context.Canceled.Error(), "context was cancelled")
		assert.EqualError(t, err, context.Canceled.Error(), "process was killed")
	})
}

func Test_parseClamOutput(t *testing.T) {
	t.Run("single file test with EICAR detection", func(t *testing.T) {
		contents, err := ioutil.ReadFile("./fixtures/clamout_single_eicar.txt")
		if err != nil {
			t.Fatal(err)
		}
		clamout := string(contents)

		results := parseClamOutput(clamout)

		assert.Len(t, results, 1, "single result returned")
		assert.Equal(t, true, results[0].Infected, "infection detected")
		assert.Equal(t, "Eicar-Signature", results[0].Detection, "EICAR signature found")
		assert.Equal(t, "/eicar/eicar", results[0].Path)

		assert.NotEmpty(t, results[0])
	})
	t.Run("single clean file test", func(t *testing.T) {
		contents, err := ioutil.ReadFile("./fixtures/clamout_single_clean.txt")
		if err != nil {
			t.Fatal(err)
		}
		clamout := string(contents)

		results := parseClamOutput(clamout)

		assert.Len(t, results, 1, "single result returned")
		assert.Equal(t, false, results[0].Infected, "no infection detected")
		assert.Equal(t, "", results[0].Detection, "empty detection string")
		assert.Equal(t, "/eicar/clean", results[0].Path)

		assert.NotEmpty(t, results[0])
	})
}
