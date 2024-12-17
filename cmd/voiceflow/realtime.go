// cmd/voiceflow/realtime.go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/telepace/voiceflow/pkg/voiceprocessor"
)

var realtimeCmd = &cobra.Command{
	Use:   "realtime",
	Short: "在终端中实时监听语音并翻译",
	RunE:  runRealtime,
}

func init() {
	rootCmd.AddCommand(realtimeCmd)
}

func runRealtime(cmd *cobra.Command, args []string) error {
	fmt.Println("启动实时语音监听...")
	err := voiceprocessor.StartRealtime()
	if err != nil {
		return err
	}
	return nil
}
