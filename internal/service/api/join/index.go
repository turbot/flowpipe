package join

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterAPI(router *gin.RouterGroup) {
	router.POST("/join", joinPost)
}

type joinRequest struct {
	ID    string `json:"id"`
	Addr  string `json:"addr"`
	Voter bool   `json:"voter"`
}

func joinPost(c *gin.Context) {
	var req joinRequest

	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Debug("received join request from node", "id", req.ID, "address", req.Addr)

	// Confirm that this node can resolve the remote address. This can happen due
	// to incomplete DNS records across the underlying infrastructure. If it can't
	// then don't consider this join attempt successful -- so the joining node
	// will presumably try again.
	if addr, err := resolvableAddress(req.Addr); err != nil {
		slog.Error("failed to resolve address while handling join request", "address", req.Addr, "error", err)
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("can't resolve %s (%s)", addr, err.Error())})
		return
	}

	/*
		jr := &command.JoinRequest{
			Id:      req.Id,
			Address: req.Addr,
			Voter:   req.Voter,
		}
			if err := s.store.Join(jr); err != nil {
				if err == store.ErrNotLeader {
					leaderAPIAddr := s.LeaderAPIAddr()
					if leaderAPIAddr == "" {
						http.Error(w, ErrLeaderNotFound.Error(), http.StatusServiceUnavailable)
						return
					}

					redirect := s.FormRedirect(r, leaderAPIAddr)
					http.Redirect(w, r, redirect, http.StatusMovedPermanently)
					return
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
	*/

	c.JSON(200, gin.H{
		"result": "JOINED",
	})
}

func resolvableAddress(addr string) (string, error) {
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Just try the given address directly.
		h = addr
	}
	_, err = net.LookupHost(h)
	return h, err
}
