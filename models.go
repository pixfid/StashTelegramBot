package main

// GraphQLRequest StashApp GraphQL структуры
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type Scene struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
	Files []struct {
		Path string `json:"path"`
	} `json:"files"`
	Paths struct {
		Screenshot string `json:"screenshot"`
		Stream     string `json:"stream"`
		Preview    string `json:"preview"`
		Sprite     string `json:"sprite"`
	} `json:"paths"`
	Tags []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
	Performers []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"performers"`
	Studio struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"studio"`
}

type GraphQLResponse struct {
	Data struct {
		FindScenes struct {
			Scenes []Scene `json:"scenes"`
			Count  int     `json:"count"`
		} `json:"findScenes"`
		FindScene Scene `json:"findScene"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}
