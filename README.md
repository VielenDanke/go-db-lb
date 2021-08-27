# Load Balancer for sql, sqlx, pgxpool nodes

NewLoadBalancer function expects two arguments
1. nodeSize - expected amount of nodes
2. timeoutSeconds - timeout between health check ping calls

<b>If node is not exists you will receive nil value</b>

## PoolNode interface

````
type PoolNode interface {
	DBx() *sqlx.DB
	DB() *sql.DB
	PGxPool() *pgxpool.Pool
	Ping(ctx context.Context) error
	IsPrimary() bool
	IsHealthy() bool
	SetHealthyStatus(status bool)
	SetPrimaryStatus(status bool)
}
````

<b>Be aware - If you would use sqlx, you have to call DBx() function on result PoolNode, 
and the same logic applied for other implementation. 
If you wouldn't - you would receive nil value</b>

### SQL example

````
sqlLB := NewLoadBalancer(ctx, 2, 2)

dbF := sqlLB.CallPrimaryPreferred().DB()
dbS := sqlLB.CallFirstAvailable().DB()
dbT := sqlLB.CallPrimary().DB()
````

### SQLx example

````
sqxlLB := NewLoadBalancer(ctx, 2, 2)

dbF := sqlxLB.CallPrimaryPreferred().DBx()
dbS := sqlxLB.CallFirstAvailable().DBx()
dbT := sqlxLB.CallPrimary().DBx()
````

### PGxPool example

````
pgxLB := NewLoadBalancer(ctx, 2, 2)

dbF := pgxLB.CallPrimaryPreferred().PGxPool()
dbS := pgxLB.CallFirstAvailable().PGxPool()
dbT := pgxLB.CallPrimary().PGxPool()
````