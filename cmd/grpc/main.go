package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/oegegr/shortener/pkg/client"
)

func main() {
	// Всего 3 режима работы:
	// 1. grpc-client <url> - сократить URL
	// 2. grpc-client -l - показать список
	// 3. grpc-client -e <id> - расширить ID
	
	var list bool
	var expand string
	var address string
	
	flag.BoolVar(&list, "l", false, "List user URLs")
	flag.StringVar(&expand, "e", "", "Expand short URL ID")
	flag.StringVar(&address, "a", "localhost:8082", "Server address")
	
	flag.Parse()
	
	args := flag.Args()
	
	clientCfg := client.Config{
		Address: address,
		Timeout: 10 * time.Second,
	}
	
	c, err := client.NewGRPCClient(clientCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	
	ctx := context.Background()
	
	switch {
	case list:
		urls, err := c.ListUserURLs(ctx)
		if err != nil {
			log.Fatal(err)
		}
		for _, url := range urls {
			fmt.Printf("%s -> %s\n", url.ShortUrl, url.OriginalUrl)
		}
		
	case expand != "":
		result, err := c.ExpandURL(ctx, expand)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
		
	case len(args) > 0:
		url := args[0]
		result, err := c.ShortenURL(ctx, url)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
		
	default:
		fmt.Println("Usage:")
		fmt.Println("  grpc-client <url>            - Shorten URL")
		fmt.Println("  grpc-client -l               - List URLs")
		fmt.Println("  grpc-client -e <id>          - Expand short URL")
		fmt.Println("  grpc-client -a <addr> <url>  - Use specific server")
	}
}