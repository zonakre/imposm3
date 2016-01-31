package postgis

import (
	"fmt"
)

type ColumnType interface {
	Name() string
	IsGeometry() bool
	PrepareInsertSql(i int,
		spec *TableSpec) string
	GeneralizeSql(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string
}

type simpleColumnType struct {
	name string
}

func (t *simpleColumnType) Name() string {
	return t.name
}

func (t *simpleColumnType) IsGeometry() bool {
	return false
}

func (t *simpleColumnType) PrepareInsertSql(i int, spec *TableSpec) string {
	return fmt.Sprintf("$%d", i)
}

func (t *simpleColumnType) GeneralizeSql(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string {
	return "\"" + colSpec.Name + "\""
}

type hstoreColumnType struct {
	simpleColumnType
}

func (t *hstoreColumnType) PrepareInsertSql(i int, spec *TableSpec) string {
	return fmt.Sprintf("$%d::hstore", i)
}

type geometryType struct {
	name string
}

func (t *geometryType) Name() string {
	return t.name
}

func (t *geometryType) IsGeometry() bool {
	return true
}

func (t *geometryType) PrepareInsertSql(i int, spec *TableSpec) string {
	return fmt.Sprintf("$%d::Geometry",
		i,
	)
}

func (t *geometryType) GeneralizeSql(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string {
	return fmt.Sprintf(`ST_SimplifyPreserveTopology("%s", %f) as "%s"`,
		colSpec.Name, spec.Tolerance, colSpec.Name,
	)
}

type validatedGeometryType struct {
	geometryType
}

func (t *validatedGeometryType) GeneralizeSql(colSpec *ColumnSpec, spec *GeneralizedTableSpec) string {
	if spec.Source.GeometryType != "polygon" {
		// TODO return warning earlier
		log.Warnf("validated_geometry column returns polygon geometries for %s", spec.FullName)
	}
	return fmt.Sprintf(`ST_Buffer(ST_SimplifyPreserveTopology("%s", %f), 0) as "%s"`,
		colSpec.Name, spec.Tolerance, colSpec.Name,
	)
}

var pgTypes = map[string]ColumnType{
	"string":             &simpleColumnType{"VARCHAR"},
	"bool":               &simpleColumnType{"BOOL"},
	"int8":               &simpleColumnType{"SMALLINT"},
	"int32":              &simpleColumnType{"INT"},
	"int64":              &simpleColumnType{"BIGINT"},
	"float32":            &simpleColumnType{"REAL"},
	"hstore_string":      &simpleColumnType{"HSTORE"},
	"geometry":           &geometryType{"GEOMETRY"},
	"validated_geometry": &validatedGeometryType{geometryType{"GEOMETRY"}},
	"geometry_noindex":   &geometryType{"GEOMETRYNOINDEX"},
	"point":              &geometryType{"POINT"},
	"linestring":         &geometryType{"LINESTRING"},
	"json_string":        &simpleColumnType{"JSON"},  // only  >= PostgreSQL 9.2
	"jsonb_string":       &simpleColumnType{"JSONB"}, // only  >= PostgreSQL 9.4
	"date":               &simpleColumnType{"DATE"},
	"time":               &simpleColumnType{"TIME"},
	"timestamp":          &simpleColumnType{"TIMESTAMP"},
	"char1":              &simpleColumnType{"\"char\""}, // PostgreSQL single-byte internal type ;
}

func registerColumnType(ColumnTypeName string, columntype ColumnType) {

	if _, ok := pgTypes[ColumnTypeName]; ok {
		panic("postgis.RegisterColumnType duplicate key: " + ColumnTypeName)
	} else {
		pgTypes[ColumnTypeName] = columntype
	}
}
