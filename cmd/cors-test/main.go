package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rocky-ads/site/internal/config"
)

func main() {
	publicURL := flag.String("url",
		os.Getenv("MINIO_PUBLIC_URL"),
		"Public MinIO API URL (or set MINIO_PUBLIC_URL)")
	key := flag.String("key", "staging/cors-test/hello.txt",
		"Object key to presign for PUT")
	expiry := flag.Duration("expiry", 15*time.Minute,
		"Presigned URL lifetime")
	contentType := flag.String("content-type", "text/plain",
		"Content-Type the browser must send on PUT")
	verify := flag.Bool("verify", false,
		"Also PUT a tiny object via the SDK to check credentials")
	flag.Parse()

	if *publicURL == "" {
		fmt.Fprintln(os.Stderr,
			"error: -url or MINIO_PUBLIC_URL is required "+
				"(e.g. https://minio.rockyads.com)")
		os.Exit(1)
	}
	if config.MinIORootUser == "" || config.MinIORootPassword == "" {
		fmt.Fprintln(os.Stderr,
			"error: MINIO_ROOT_USER and MINIO_ROOT_PASSWORD must be set")
		os.Exit(1)
	}
	if config.MinIOBucketName == "" {
		fmt.Fprintln(os.Stderr,
			"error: MINIO_BUCKET_NAME must be set")
		os.Exit(1)
	}

	client, err := newPublicClient(*publicURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	bucket := config.MinIOBucketName

	if *verify {
		_, err := client.PutObject(ctx, bucket, *key,
			strings.NewReader("cors-test"), 9,
			minio.PutObjectOptions{ContentType: *contentType})
		if err != nil {
			fmt.Fprintf(os.Stderr, "verify PUT failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("verify: wrote s3://%s/%s via SDK\n", bucket, *key)
	}

	presigned, err := client.PresignedPutObject(ctx, bucket, *key, *expiry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "presign failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("bucket:", bucket)
	fmt.Println("key:", *key)
	fmt.Println("expiry:", *expiry)
	fmt.Println("content-type:", *contentType)
	fmt.Println()
	fmt.Println("presigned PUT URL:")
	fmt.Println(presigned.String())
	fmt.Println()
	fmt.Println(browserPasteHint(*publicURL))
	fmt.Println()
	fmt.Print(browserSnippet(presigned.String(), *contentType))
}

func browserPasteHint(rawURL string) string {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "Paste this in the browser console on a page " +
			"allowed by your MinIO bucket CORS rules:"
	}
	host := strings.ToLower(u.Hostname())
	if host == "127.0.0.1" || host == "localhost" {
		return "Local MinIO: run this in the console of a page " +
			"served from your machine (e.g. http://localhost:<app-port>), " +
			"not https://rockyads.com. That origin must be in bucket CORS. " +
			"Mixed-content blocks http:// MinIO from https:// pages."
	}
	return "Paste this in the browser console on a page whose " +
		"origin is allowed by MinIO CORS for this host " +
		"(e.g. https://rockyads.com → https://minio.rockyads.com)."
}

func newPublicClient(rawURL string) (*minio.Client, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("URL missing host: %q", rawURL)
	}

	client, err := minio.New(u.Host, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.MinIORootUser, config.MinIORootPassword, ""),
		Secure: u.Scheme == "https",
	})
	if err != nil {
		return nil, fmt.Errorf("initialize MinIO client: %w", err)
	}
	return client, nil
}

func browserSnippet(putURL, contentType string) string {
	return fmt.Sprintf(`const url = %q;
const blob = new Blob(["hello from browser"], { type: %q });
const xhr = new XMLHttpRequest();
xhr.upload.onprogress = (e) => {
  if (e.lengthComputable) console.log("progress", e.loaded, "/", e.total);
};
xhr.open("PUT", url);
xhr.setRequestHeader("Content-Type", %q);
xhr.onload = () => console.log("status", xhr.status, xhr.responseText);
xhr.onerror = () => console.error("network/CORS error", xhr.status);
xhr.send(blob);
`, putURL, contentType, contentType)
}
