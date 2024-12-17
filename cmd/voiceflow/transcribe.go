// cmd/voiceflow/transcribe.go
package main

//import (
//	"fmt"
//	"github.com/spf13/cobra"
//	"github.com/telepace/voiceflow/pkg/voiceprocessor"
//)
//
//var transcribeCmd = &cobra.Command{
//	Use:   "transcribe [音频文件路径]",
//	Short: "转录并翻译指定的音频文件",
//	Args:  cobra.ExactArgs(1),
//	RunE:  runTranscribe,
//}
//
//func init() {
//	rootCmd.AddCommand(transcribeCmd)
//}
//
//func runTranscribe(cmd *cobra.Command, args []string) error {
//	audioFile := args[0]
//	fmt.Printf("正在转录音频文件：%s\n", audioFile)
//	err := voiceprocessor.TranscribeFile(audioFile)
//	if err != nil {
//		return err
//	}
//	return nil
//}
