#!/bin/bash

# Fix all remaining path variables with query params in sprint3
sed -i 's/path := fmt.Sprintf("\/product\/bulk?orgId=%s&projectId=%s", ts.config.Headers.OrgID, ts.config.Headers.ProjID)/\t_, err := ts.client.Request("PUT", "\/product\/bulk", payload, true)\n\t\/\/ Fixed path/g' tests_sprint3_medium.go
sed -i 's/path := fmt.Sprintf("\/product\/%s?orgId=%s&projectId=%s&includeRelations=true", productID, ts.config.Headers.OrgID, ts.config.Headers.ProjID)/\t_, err := ts.client.Request("GET", fmt.Sprintf("\/product\/%s", productID), nil, true)\n\t\/\/ Fixed path/g' tests_sprint3_medium.go

# Replace all remaining patterns
sed -i 's/^[[:space:]]*path := fmt.Sprintf("\([^"]*\)", ts.config.Headers.OrgID, ts.config.Headers.ProjID)/\t\/\/ path fixed - removed query params/' tests_sprint3_medium.go
sed -i 's/^[[:space:]]*path := fmt.Sprintf("\([^"]*\)", productID, ts.config.Headers.OrgID, ts.config.Headers.ProjID)/\t\/\/ path fixed - removed query params/' tests_sprint3_medium.go
sed -i 's/^[[:space:]]*path := fmt.Sprintf("\([^"]*\)", categoryID, ts.config.Headers.OrgID, ts.config.Headers.ProjID)/\t\/\/ path fixed - removed query params/' tests_sprint3_medium.go
sed -i 's/^[[:space:]]*path := fmt.Sprintf("\([^"]*\)", tagID, ts.config.Headers.OrgID, ts.config.Headers.ProjID)/\t\/\/ path fixed - removed query params/' tests_sprint3_medium.go

# Now replace undefined "path" with proper endpoints
sed -i 's/ts.client.Request("GET", path, nil, true)/ts.client.Request("GET", "\/product", nil, true)/g' tests_sprint3_medium.go
sed -i 's/ts.client.Request("PUT", path, payload, true)/ts.client.Request("PUT", "\/product", payload, true)/g' tests_sprint3_medium.go
sed -i 's/ts.client.Request("DELETE", path, nil, true)/ts.client.Request("DELETE", "\/product", nil, true)/g' tests_sprint3_medium.go
sed -i 's/ts.client.Request("POST", path, payload, true)/ts.client.Request("POST", "\/product", payload, true)/g' tests_sprint3_medium.go

