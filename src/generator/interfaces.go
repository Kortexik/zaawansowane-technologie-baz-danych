package generator

type QueryGenerator interface {
	SQLInsert() string
	SQLSelect(id int) string
	SQLUpdate(id int, field string, value any) string
	SQLDelete(id int) string

	MongoInsert() string
	MongoFind(id int) string
	MongoUpdate(id int, field string, value any) string
	MongoDelete(id int) string

	RedisSet() string
	RedisGet(id int) string
	RedisDelete(id int) string
}
