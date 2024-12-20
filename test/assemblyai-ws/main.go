package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/gordonklaus/portaudio"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	checkErr(err)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// We need portaudio to record the microphone.
	err = portaudio.Initialize()
	checkErr(err)
	defer portaudio.Terminate()

	var (
		// Number of samples per seconds.
		sampleRate      = 16_000
		framesPerBuffer = 3_200
	)

	transcriber := &aai.RealTimeTranscriber{
		OnSessionBegins: func(event aai.SessionBegins) {
			fmt.Println("session begins")
		},
		OnSessionTerminated: func(event aai.SessionTerminated) {
			fmt.Println("session terminated")
		},
		OnFinalTranscript: func(transcript aai.FinalTranscript) {
			fmt.Println(transcript.Text)
		},
		OnPartialTranscript: func(transcript aai.PartialTranscript) {
			fmt.Printf("%s\r", transcript.Text)
		},
		OnError: func(err error) {
			fmt.Printf("Something bad happened: %v", err)
		},
	}

	apiKey := os.Getenv("ASSEMBLYAI_API_KEY")

	client := aai.NewRealTimeClientWithOptions(
		aai.WithRealTimeAPIKey(apiKey),
		aai.WithRealTimeSampleRate(int(sampleRate)),
		aai.WithRealTimeTranscriber(transcriber),
	)

	ctx := context.Background()

	err = client.Connect(ctx)
	checkErr(err)

	slog.Info("connected to real-time API", "sample_rate", sampleRate, "frames_per_buffer", framesPerBuffer)

	rec, err := NewRecorder(sampleRate, framesPerBuffer)
	checkErr(err)

	err = rec.Start()
	checkErr(err)

	slog.Info("recording...")

	for {
		select {
		case <-sigs:
			slog.Info("stopping recording...")

			var err error

			err = rec.Stop()
			checkErr(err)

			err = client.Disconnect(ctx, true)
			checkErr(err)

			os.Exit(0)
		default:
			b, err := rec.Read()
			checkErr(err)

			// Send partial audio samples.
			err = client.Send(ctx, b)
			checkErr(err)
		}
	}
}

func checkErr(err error) {
	if err != nil {
		slog.Error("Something bad happened", "err", err)
		os.Exit(1)
	}
}
