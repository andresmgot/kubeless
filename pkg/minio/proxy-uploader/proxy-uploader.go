/*
Serve is a very simple static file server in go
Usage:
	-p="8100": port to serve on
	-d=".":    the directory of static files to host
Navigating to http://localhost:8100 will display the index.html or directory
listing file.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/minio/minio-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func uploadToMinio(file string, minioCli *minio.Client) error {
	// Upload the zip file
	objectName := path.Base(file)
	filePath := file
	c, err := ioutil.ReadFile(file)
	contentType := http.DetectContentType(c)

	// Upload the file with FPutObject
	n, err := minioCli.FPutObject("functions", objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	fmt.Printf("Successfully uploaded %s of size %d\n", objectName, n)
	return nil
}

func getMinioClient(namespace string, cli kubernetes.Interface) (*minio.Client, error) {
	minioSVC, err := cli.CoreV1().Services(namespace).Get("minio", metav1.GetOptions{})
	if err != nil {
		return &minio.Client{}, err
	}
	minioSecret, err := cli.CoreV1().Secrets(namespace).Get("minio-key", metav1.GetOptions{})
	if err != nil {
		return &minio.Client{}, err
	}
	endpoint := "minio." + namespace + ":" + strconv.Itoa(int(minioSVC.Spec.Ports[0].Port))
	accessKeyID := string(minioSecret.Data["accesskey"])
	secretAccessKey := string(minioSecret.Data["secretkey"])
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return &minio.Client{}, err
	}
	return minioClient, nil
}

func main() {
	port := flag.String("p", "8080", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()
	namespace := "default"
	if os.Getenv("MINIO_NAMESPACE") != "" {
		namespace = os.Getenv("MINIO_NAMESPACE")
	}
	cli := utils.GetClient()
	minioClient, err := getMinioClient(namespace, cli)
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.Dir(*directory)))
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		dest := path.Join(*directory, path.Base(handler.Filename))
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		n, err := io.Copy(f, file)
		if err != nil {
			panic(err)
		}
		fmt.Printf("File %s successfully stored. Written %s bytes\n", dest, strconv.Itoa(int(n)))
		err = uploadToMinio(dest, minioClient)
		if err != nil {
			panic(err)
		}
	})

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
