package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/service/fsm"
)

func (api *APIService) playRegister(router *gin.RouterGroup) {
	router.GET("/play", api.playStats)
	router.POST("/play", api.playPost)
	router.GET("/play/:key", api.playGet)
}

func (api *APIService) playPost(c *gin.Context) {
	var kv fsm.KeyValueOperation
	if err := c.BindJSON(&kv); err != nil {
		return
	}
	kv.Operation = "set"
	kvj, _ := json.Marshal(kv)
	f := api.raftService.Raft.Apply(kvj, time.Second)
	if err := f.Error(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"commit_index": f.Index(),
	})
}

func (api *APIService) playGet(c *gin.Context) {
	key := c.Param("key")
	v, err := api.raftService.Storage.Get(key)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"key":   key,
		"value": v,
	})
}

func (api *APIService) playStats(c *gin.Context) {
	c.JSON(200, gin.H{"now": time.Now()})
}

/*


func RegisterPublicAPI(router *gin.RouterGroup) {
	router.POST("/play", playPost)
	router.GET("/play/:key", playGet)
}

func playPost(c *gin.Context) {

	var kv playData

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&kv); err != nil {
		return
	}

	kv.Operation = "create"

	kvj, _ := json.Marshal(kv)

	f := r.raft.Apply(kvj, time.Second)
	if err := f.Error(); err != nil {
		return nil, rafterrors.MarkRetriable(err)
	}

	c.JSON(200, gin.H{
		"commit_index": f.Index(),
	})

}

func playGet(c *gin.Context) {
	key := c.Param("key")

	c.JSON(200, gin.H{
		"key": key,
	})

}

*/
/*


func (r rpcInterface) GetWords(ctx context.Context, req *pb.GetWordsRequest) (*pb.GetWordsResponse, error) {
	r.wordTracker.mtx.RLock()
	defer r.wordTracker.mtx.RUnlock()
	return &pb.GetWordsResponse{
		BestWords:   cloneWords(r.wordTracker.words),
		ReadAtIndex: r.raft.AppliedIndex(),
	}, nil
}


*/
