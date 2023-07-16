package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type TTSConfig struct {
	Client      *texttospeech.Client
	AudioConfig *texttospeechpb.AudioConfig
	Voice       *texttospeechpb.VoiceSelectionParams
}

func prepareAudio() *TTSConfig {
	ctx := context.Background()

	// Create a new TTS client.
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Build the voice request, select the language code ("de-DE") and the SSML voice gender.
	voice := &texttospeechpb.VoiceSelectionParams{
		LanguageCode: "de-DE",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
	}

	// Select the type of audio file you want returned.
	audioConfig := &texttospeechpb.AudioConfig{
		AudioEncoding: texttospeechpb.AudioEncoding_MP3,
	}

	return &TTSConfig{
		Client:      client,
		AudioConfig: audioConfig,
		Voice:       voice,
	}
}

func makeAudio(config *TTSConfig, text string) []byte {
	ctx := context.Background()

	// Set the text input to be synthesized.
	input := &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
	}

	// Perform the text-to-speech request on the text input with the selected voice parameters and audio file type.
	response, err := config.Client.SynthesizeSpeech(ctx, &texttospeechpb.SynthesizeSpeechRequest{
		Input:       input,
		Voice:       config.Voice,
		AudioConfig: config.AudioConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	return response.AudioContent
}

func main() {
	config := prepareAudio()

	scanner := bufio.NewScanner(os.Stdin)
	re := regexp.MustCompile(`\((.*)\)`)
	reSpecial := regexp.MustCompile(`[.,!?]`)

	// Open output.txt for writing.
	f, err := os.Create("output.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		var translation string
		if len(matches) > 1 {
			translation = strings.TrimSpace(matches[1]) // Extract translation.
		}
		phrase := re.ReplaceAllString(line, "") // Remove translation in brackets.
		phrase = strings.TrimSpace(phrase)      // Remove leading and trailing white space.
		filename := reSpecial.ReplaceAllString(phrase, "")
		filename = strings.ReplaceAll(filename, " ", "-")
		filename += ".mp3"

		audio := makeAudio(config, phrase)
		err := ioutil.WriteFile(filename, audio, 0o644)
		if err != nil {
			log.Fatal(err)
		}

		// Write the original phrase and translation to output.txt.
		_, err = f.WriteString(fmt.Sprintf("%s\t%s\n", phrase, translation))
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
