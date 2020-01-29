package rununtil_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/kaluza-tech/rununtil"
)

func helperSendSignal(t *testing.T, p *os.Process, sent *bool, signal os.Signal, delay time.Duration) {
	time.Sleep(delay)
	if err := p.Signal(signal); err != nil {
		t.Errorf("unexpected error occurred: %v", err)
	}
	*sent = true
}

func helperMakeFakeRunner(hasBeenShutdown *bool) rununtil.RunnerFunc {
	return rununtil.RunnerFunc(func() rununtil.ShutdownFunc {
		return rununtil.ShutdownFunc(func() {
			*hasBeenShutdown = true
		})
	})
}

func helperMakeMain(hasBeenKilled *bool) func() {
	return func() {
		rununtil.AwaitKillSignal(helperMakeFakeRunner(hasBeenKilled))
	}
}

func TestRununtilAwaitKillSignal(t *testing.T) {
	table := []struct {
		name   string
		signal os.Signal
	}{
		{
			name:   "Shutdown from SIGINT",
			signal: syscall.SIGINT,
		},
		{
			name:   "Shutdown from SIGTERM",
			signal: syscall.SIGTERM,
		},
	}
	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			var sentSignal bool
			var hasBeenShutdown bool
			p, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatalf("Unexpected error when finding process: %v", err)
			}

			go helperSendSignal(t, p, &sentSignal, test.signal, 1*time.Millisecond)
			rununtil.AwaitKillSignal(helperMakeFakeRunner(&hasBeenShutdown))
			if !sentSignal {
				t.Fatal("expected signal to have been sent")
			}
			if !hasBeenShutdown {
				t.Fatal("expected the shutdown function to have been called")
			}
		})
	}
}

func TestRununtilAwaitKillSignal_MultipleRunnerFuncs(t *testing.T) {
	var hasBeenShutdown1, hasBeenShutdown2, hasBeenShutdown3 bool
	var sentSignal bool

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Unexpected error when finding process: %v", err)
	}

	go helperSendSignal(t, p, &sentSignal, syscall.SIGINT, time.Millisecond)

	rununtil.AwaitKillSignal(
		helperMakeFakeRunner(&hasBeenShutdown1),
		helperMakeFakeRunner(&hasBeenShutdown2),
		helperMakeFakeRunner(&hasBeenShutdown3),
	)

	if !sentSignal {
		t.Fatal("expected signal to have been sent")
	}
	if !hasBeenShutdown1 {
		t.Fatal("expected the shutdown function 1 to have been called")
	}
	if !hasBeenShutdown2 {
		t.Fatal("expected the shutdown function 2 to have been called")
	}
	if !hasBeenShutdown3 {
		t.Fatal("expected the shutdown function 3 to have been called")
	}
}

func TestRununtilKilled(t *testing.T) {
	var hasBeenKilled bool
	cancel := rununtil.Killed(helperMakeMain(&hasBeenKilled))
	cancel()

	// yield control back to scheduler so that killing can actually happen
	time.Sleep(time.Millisecond)
	if !hasBeenKilled {
		t.Fatal("expected main to have been killed")
	}
}

func TestRununtilCancelAll(t *testing.T) {
	var hasBeenKilled bool
	rununtil.Killed(helperMakeMain(&hasBeenKilled))

	// yield control back to scheduler so that the go routines can actually
	// start
	time.Sleep(time.Millisecond)

	rununtil.CancelAll()

	// yield control back to scheduler so that killing can actually happen
	time.Sleep(time.Millisecond)
	if !hasBeenKilled {
		t.Fatal("expected main to have been killed")
	}
}

func TestRununtilCancelAll_MultipleTimes(t *testing.T) {
	var hasBeenKilled bool
	for idx := 0; idx < 100; idx++ {
		hasBeenKilled = false
		rununtil.Killed(helperMakeMain(&hasBeenKilled))

		// yield control back to scheduler so that the go routines can actually
		// start
		time.Sleep(time.Millisecond)

		rununtil.CancelAll()

		// yield control back to scheduler so that killing can actually happen
		time.Sleep(time.Millisecond)
		if !hasBeenKilled {
			t.Fatal("expected main to have been killed")
		}
	}
}

func TestRununtilCancelAll_Threadsafe(t *testing.T) {
	var hasBeenKilledVec [100]bool
	for idx := 0; idx < 100; idx++ {
		cancel := rununtil.Killed(helperMakeMain(&hasBeenKilledVec[idx]))
		cancel()
		rununtil.CancelAll()
	}
	// yield control back to scheduler so that killing can actually happen
	time.Sleep(time.Millisecond)
	for idx, hasBeenKilled := range hasBeenKilledVec {
		if !hasBeenKilled {
			t.Fatalf("expected main to have been killed: %d", idx)
		}
	}
}

// Annoyingly this test has to be run by itself to actually fail...
//	go test -v -run TestKilled_FailsForNonblockingMain
// Fixed test by not actually sending a kill signal anymore --
// it now calls rununtil.CancelAll().
func TestKilled_FailsForNonblockingMain(t *testing.T) {
	cancel := rununtil.Killed(func() {})
	cancel()
}
