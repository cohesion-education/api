package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cohesion-education/admin-api/pkg/cohesioned/auth"
	"github.com/cohesion-education/admin-api/pkg/cohesioned/config"
	"github.com/cohesion-education/admin-api/pkg/cohesioned/gcp"
	"github.com/cohesion-education/admin-api/pkg/cohesioned/homepage"
	"github.com/cohesion-education/admin-api/pkg/cohesioned/profile"
	"github.com/cohesion-education/admin-api/pkg/cohesioned/taxonomy"
	"github.com/cohesion-education/admin-api/pkg/cohesioned/video"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var (
	apiRenderer = render.New()
)

func Run(port string) {
	s := newServer()
	s.Run(":" + port)
}

func newServer() *negroni.Negroni {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Failed to load .env file. Components will fallback to loading from VCAP_SERVICES or env vars")
	}

	authConfig, err := config.NewAuthConfig()
	if err != nil {
		log.Fatal(err)
	}

	gcpKeyfileLocation := os.Getenv("GCP_KEYFILE_LOCATION")
	if len(gcpKeyfileLocation) == 0 {
		log.Fatal("Required env var GCP_KEYFILE_LOCATION not set")
	}

	videoStorageBucketName := os.Getenv("GCP_STORAGE_VIDEO_BUCKET")
	if len(videoStorageBucketName) == 0 {
		log.Fatal("Required env var GCP_STORAGE_VIDEO_BUCKET not set")
	}

	gcpConfig, err := gcp.NewConfig(gcpKeyfileLocation)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.TODO()
	datastoreClient, err := gcp.NewDatastoreClient(ctx, gcpConfig)
	if err != nil {
		log.Fatalf("Failed to get datastore client %v", err)
	}

	storageClient, err := gcp.NewStorageClient(ctx, gcpConfig)
	if err != nil {
		log.Fatalf("Failed to get storage client %v", err)
	}

	homepageRepo := homepage.NewGCPDatastoreRepo(datastoreClient)
	profileRepo := profile.NewGCPDatastoreRepo(datastoreClient)
	taxonomyRepo := taxonomy.NewGCPDatastoreRepo(datastoreClient)
	videoRepo := video.NewGCPRepo(datastoreClient, storageClient, videoStorageBucketName)

	n := negroni.Classic()
	mx := mux.NewRouter()
	mx.StrictSlash(true)

	// Static Assets
	mx.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/assets/"))))
	mx.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/build/static/"))))

	//TODO - register 404
	// mx.NotFoundHandler = cohesioned.NotFoundViewHandler(homepageRenderer)

	//Public APIs
	mx.Methods(http.MethodGet).Path("/api/homepage").Handler(homepage.HomepageHandler(apiRenderer, homepageRepo))
	mx.Methods(http.MethodGet).Path("/api/taxonomy/flatten").Handler(taxonomy.FlatListHandler(apiRenderer, taxonomyRepo))

	isAuthenticatedHandler := auth.IsAuthenticatedHandler(authConfig)
	authMiddleware := negroni.New(
		negroni.HandlerFunc(isAuthenticatedHandler),
	)

	//APIs that require Admin priveleges
	requiresAdmin(http.MethodPost, "/api/taxonomy", taxonomy.AddHandler(apiRenderer, taxonomyRepo), mx, authMiddleware)
	requiresAdmin(http.MethodPost, "/api/video", video.SaveHandler(apiRenderer, videoRepo), mx, authMiddleware)
	requiresAdmin(http.MethodPut, "/api/video", video.UpdateHandler(apiRenderer, videoRepo), mx, authMiddleware)

	//APIs that only require Authentication
	requiresAuth(http.MethodPost, "/api/profile/preferences", profile.SavePreferencesHandler(apiRenderer, authConfig, profileRepo), mx, authMiddleware)
	requiresAuth(http.MethodGet, "/api/taxonomy", taxonomy.ListHandler(apiRenderer, taxonomyRepo), mx, authMiddleware)
	requiresAuth(http.MethodGet, "/api/taxonomy/{id:[0-9]+}/children", taxonomy.ListChildrenHandler(apiRenderer, taxonomyRepo), mx, authMiddleware)
	requiresAuth(http.MethodGet, "/api/video/stream/{id:[0-9]+}", video.StreamHandler(apiRenderer, videoRepo, gcpConfig), mx, authMiddleware)

	// mx.PathPrefix("/callback").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 	fmt.Println("serving /callback as index.html")
	// 	http.ServeFile(w, req, "./web/build/index.html")
	// }))

	mx.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("serving %s as index.html\n", req.RequestURI)
		http.ServeFile(w, req, "./web/build/index.html")
	}))

	n.UseHandler(mx)
	return n
}

func requiresAuth(method string, uri string, handler http.Handler, mx *mux.Router, authMiddleware *negroni.Negroni) {
	mx.Methods(method).Path(uri).Handler(authMiddleware.With(
		negroni.Wrap(handler),
	))
}

func requiresAdmin(method string, uri string, handler http.Handler, mx *mux.Router, authMiddleware *negroni.Negroni) {
	mx.Methods(method).Path(uri).Handler(authMiddleware.With(
		negroni.HandlerFunc(auth.IsAdmin),
		negroni.Wrap(handler),
	))
}