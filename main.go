package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	log.Println(os.Args)
	if len(os.Args) != 2 {
		log.Printf("need port number\n")
		os.Exit(1)
	}
	p := os.Args[1]
	log.Printf("port:%s", p)
	l, err := net.Listen("tcp", ":"+p)
	if err != nil {
		return
	}
	err = run(context.Background(), l)
	//if err != nil {
	//	log.Printf("failed to terminate server: %v", err)
	//}
}

func run(ctx context.Context, l net.Listener) error {
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
		}),
	}
	eg, ctx := errgroup.WithContext(ctx)

	// 別ゴルーチンでHTTPサーバーを起動する
	eg.Go(func() error {
		// http.ErrServerClosedはhttp.Server.Shutdown()が正常に終了したことを示すので異常ではないので除外
		err := server.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("failed to close: %+v", err)
			return err
		}
		return nil
	})

	// チャネルからの通知を待機する
	<-ctx.Done()
	log.Printf("ctx done")
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}

	// Goメソッドで起動した別ゴルーチンの終了をまつ
	return eg.Wait()
}
