/*
Copyright 2019 The Jetstack cert-manager contributors.
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"context"
	"flag"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var (
	Log = klogr.New().WithName("e2e-metrics")

	ErrorLevel = 0
	WarnLevel  = 1
	InfoLevel  = 2
	DebugLevel = 3
)

var logFlushFreq = flag.Duration(
	"log-flush-frequency",
	5*time.Second,
	"Maximum number of seconds between log flushes",
)

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	klog.Info(string(data))
	return len(data), nil
}

// InitLogs initializes logs the way we want for kubernetes.
//func InitLogs(fs *flag.FlagSet) {
// if fs == nil {
// 	fs = flag.CommandLine
// }
func InitLogs() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})

	// klog.InitFlags(fs)
	// fs.Set("logtostderr", "true")

	// log.SetOutput(GlogWriter{})
	// log.SetFlags(0)

	// The default glog flush interval is 30 seconds, which is frighteningly long.
	go wait.Until(klog.Flush, *logFlushFreq, wait.NeverStop)
}

// FlushLogs flushes logs immediately.
func FlushLogs() {
	klog.Flush()
}

// key used to lookup logger instance
var contextKey = &struct{}{}

func FromContext(ctx context.Context, names ...string) logr.Logger {
	l := ctx.Value(contextKey)
	if l == nil {
		return Log
	}
	lT := l.(logr.Logger)
	for _, n := range names {
		lT = lT.WithName(n)
	}
	return lT
}

func NewContext(ctx context.Context, l logr.Logger, names ...string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = FromContext(ctx)
	}
	for _, n := range names {
		l = l.WithName(n)
	}
	return context.WithValue(ctx, contextKey, l)
}
