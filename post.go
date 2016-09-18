package main

import (
	// basic imports
	"log"
	"os"
	"strings"
	"time"

	// extra fun imports
	"encoding/base64"
	"io/ioutil"
	"math/rand"

	// config file library
	"github.com/BurntSushi/toml"

	// tumblr go client library
	"github.com/MariaTerzieva/gotumblr"
)

type Config struct {
	Source            string
	Archive           string
	TumblrBlog        string
	TumblrConsumerKey string
	TumblrSecretKey   string
	TumblrToken       string
	TumblrTokenSecret string
	TumblrTags        []string
}

// getImageBase64 gets the image and returns a string representing that image's
// base64 encoding
func getImageBase64(config Config, image string) string {
	rawImage, err := ioutil.ReadFile(config.Source + image)
	if err != nil {
		log.Fatalf("failed to read raw image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(rawImage)
}

func tagsToString(config Config) string {
	return strings.Join(config.TumblrTags, ",")
}

func postToTumblr(config Config, image string) {
	image64 := getImageBase64(config, image)

	client := gotumblr.NewTumblrRestClient(config.TumblrConsumerKey, config.TumblrSecretKey, config.TumblrToken, config.TumblrTokenSecret, "", "http://api.tumblr.com")

	opts := map[string]string{
		"data64":  image64,
		"tags":    tagsToString(config),
		"caption": image,
	}

	if err := client.CreatePhoto(config.TumblrBlog, opts); err != nil {
		log.Fatalf("failed post to tumblr: %v", err)
	}
}

// pickImage lists the directory, picks one of the files, and return its filename
func pickImage(config Config) string {
	// get list of files
	files, err := ioutil.ReadDir(config.Source)
	if err != nil {
		log.Fatalf("cannot read directory: %v", err)
	}
	// pick one at random
	image := files[rand.Intn(len(files))].Name()

	// TODO(dperny): verify that the file is an image before we post it
	// TODO(dperny): handle the case where we have exhuasted the directory

	// return it
	return image
}

// postImage takes a directory and filename and posts it to the various apis
func postImage(config Config, image string) {
	log.Printf("posting %v", config.Source+image)
	// call out to various APIs
	postToTumblr(config, image)
	// postToTwitter(image)
	// postToFacebook(image)
}

// archiveImage takes the filename and moves it to the archive/ subdirectory,
// where it will not be picked again
func archiveImage(config Config, image string) {
	err := os.Rename(config.Source+image, config.Archive+image)
	if err != nil {
		log.Fatalf("failed to move image to archive: %v", err)
	}
}

// getConfig fetches the config from the location provided and returns it
func getConfig(configLocation string) Config {
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("failure reading config file: %v", err)
	}

	// read the config file
	var config Config
	if _, err := toml.Decode(string(data), &config); err != nil {
		log.Fatalf("failure parsing config file: %v", err)
	}
	return config
}

func main() {
	// get the config
	config := getConfig("config.toml")
	// set a random seed
	rand.Seed(time.Now().Unix())
	// pick an image
	image := pickImage(config)
	// post the image
	postImage(config, image)
	// archive the image
	archiveImage(config, image)
}
