package main

import (
	"bufio"
	"flag"
	"github.com/gpestana/redonion/fetcher"
	"github.com/gpestana/redonion/outputs"
	"github.com/gpestana/redonion/processors"
	"log"
	"os"
	"strings"
)

func main() {

	urls := flag.String("urls", "http://127.0.0.1", "list of addresses to scan (separated by comma)")
	list := flag.String("list", "", "path for list of addresses to scan")
	timeout := flag.Int("timeout", 15, "requests timeout (seconds)")
	proxy := flag.String("proxy", "127.0.0.1:9150", "url of tor proxy")
	flag.Parse()

	ulist, err := parseUrls(urls, list)
	if err != nil {
		log.Fatal(err)
	}

	//textChn := make(chan processor.DataUnit, len(ulist))
	imgChn := make(chan processor.DataUnit, len(ulist))
	//chs := []chan processor.DataUnit{textChn, imgChn}
	chs := []chan processor.DataUnit{imgChn}

	outputChn := make(chan processor.DataUnit, len(ulist)*len(chs))
	output := output.NewStdout(outputChn, len(ulist))

	processors := []processor.Processor{
		//processor.NewTextProcessor(textChn, outputChn, len(ulist)),
		processor.NewImageProcessor(imgChn, outputChn, len(ulist)),
	}

	fetcher, err := fetcher.New(ulist, proxy, timeout, processors)
	if err != nil {
		log.Fatal(err)
	}

	fetcher.Start()
	for _, p := range processors {
		p.Process()
	}
	output.Run()

	jres, err := output.Result()
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(jres)

	closeChannels(chs, outputChn)
}

func parseUrls(urlsIn *string, list *string) ([]string, error) {
	var urls []string
	if *list != "" {
		urlsPath, err := parseListURL(*list)
		if err != nil {
			return nil, err
		}
		urls = urlsPath
	} else {
		urls = strings.Split(*urlsIn, ",")
	}
	return urls, nil
}

func parseListURL(p string) ([]string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		urls = append(urls, s.Text())
	}
	return urls, s.Err()
}

func closeChannels(chs []chan processor.DataUnit, outputCh chan processor.DataUnit) {
	for _, ch := range chs {
		close(ch)
	}
	close(outputCh)
}
