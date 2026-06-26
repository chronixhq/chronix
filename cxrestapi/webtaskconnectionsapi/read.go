package webtaskconnectionsapi

import (
	"chronix/cxrestapi/apiutil"
	"chronix/internal/db"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func listWebtaskConnections(c *gin.Context) {
	rows, err := db.WebtaskConnection.Order(db.WebtaskConnection.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to list connections", err.Error())
		return
	}

	uuids := make([]string, 0)
	uuidSeen := make(map[string]struct{})
	for _, it := range rows {
		if it.AgentUUID != nil {
			u := strings.TrimSpace(*it.AgentUUID)
			if u != "" {
				if _, ok := uuidSeen[u]; !ok {
					uuidSeen[u] = struct{}{}
					uuids = append(uuids, u)
				}
			}
		}
	}
	agentNames := make(map[string]string)
	if len(uuids) > 0 {
		if agents, err := db.Agent.Where(db.Agent.UUID.In(uuids...)).Find(); err == nil {
			for _, a := range agents {
				agentNames[a.UUID] = a.Name
			}
		}
	}

	resp := make([]gin.H, 0, len(rows))
	for _, it := range rows {
		m := MapWebtaskConnection(it)
		if it.AgentUUID != nil {
			if name, ok := agentNames[strings.TrimSpace(*it.AgentUUID)]; ok {
				m["agentName"] = name
			}
		}
		resp = append(resp, m)
	}
	restresponse.RestSuccess(c, resp)
}

func getWebtaskConnection(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	row, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "connection not found")
		return
	}

	m := MapWebtaskConnection(row)
	if row.AgentUUID != nil {
		if agent, err := db.Agent.Where(db.Agent.UUID.Eq(strings.TrimSpace(*row.AgentUUID))).First(); err == nil {
			m["agentName"] = agent.Name
		}
	}
	restresponse.RestSuccess(c, m)
}
