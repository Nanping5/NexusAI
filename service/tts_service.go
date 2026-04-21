package service

import (
	"NexusAi/common/tts"
	"context"
)

type ttsService struct {
}

var TTSService = &ttsService{}

// SynthesizeSpeech 语音合成（返回文件路径）
func (s *ttsService) SynthesizeSpeech(ctx context.Context, text string) (string, error) {
	ttsClient := tts.GetTTSService()
	return ttsClient.SynthesizeAsync(ctx, text)
}

// SynthesizeSpeechData 语音合成（返回音频数据）
func (s *ttsService) SynthesizeSpeechData(ctx context.Context, text string) ([]byte, error) {
	ttsClient := tts.GetTTSService()
	return ttsClient.Synthesize(ctx, text)
}
