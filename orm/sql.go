package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var NotFound = fmt.Errorf("not found")

type Field interface {
	// Initialize the field
	Init()
	// Generate a default value
	Generate() error
	// Return a value for insertion into the database
	Get() interface{}
	// Set the new value
	Set(interface{}) error
}

type QueryModel interface {
	Database() func() DB
	SetDatabase(func() DB)
	TableName() string
	SetTableName(string)
	Init() error
}

type UpdateModel interface {
	UpdateField(string) error
}

type UpdateField interface {
	Update() error
}

type Tag struct {
	Name  string
	Value string
	Flag  bool
}

type ModelSchema struct {
	Optional            bool
	IncludeOptional     bool
	TableName           string
	Type                reflect.Type
	Fields              []*ModelSchemaField
	RelatedModelSchemas []*RelatedModelSchema
}

// describes a related model of a given model
// contains information about the foreign key
type RelatedModelSchema struct {
	Column      string
	FkField     string
	Field       string
	Optional    bool
	ModelSchema *ModelSchema
}

type ModelSchemaField struct {
	Name     string
	Column   string
	Optional bool
	Type     reflect.Type
	Value    reflect.Value
	Tags     []Tag
}

// A SelectBuffer is used to buffer values from a SELECT for later processing.
type SelectBuffer struct {
	Value interface{}
	Ptr   interface{}
	// if the buffer belongs to a field, the field value will be set here
	// so we can properly initialize it
	Field Field
}

func (s *SelectBuffer) Init() error {
	if s.Field != nil {
		return s.Field.Set(s.Value)
	}
	return nil
}

func (s *ModelSchema) New(db func() DB) QueryModel {
	model := New(s.Type)
	Init(model, db)
	return model
}

// Implements a nullable pointer
type Nullable struct {
	Ptr interface{}
}

func (n *Nullable) Scan(src interface{}) error {

	if src == nil {
		return nil
	}

	dstType := reflect.TypeOf(n.Ptr)
	dstValue := reflect.ValueOf(n.Ptr)
	srcType := reflect.TypeOf(src)
	srcValue := reflect.ValueOf(src)

	if srcType.AssignableTo(dstType.Elem()) {
		// we can directly assign the value
		dstValue.Elem().Set(srcValue)
		return nil
	} else if scanner, ok := n.Ptr.(sql.Scanner); ok {
		// we can use the scanner interface
		return scanner.Scan(src)
	}
	// we cannot assign the value
	return fmt.Errorf("cannot assign value")

}

func (s *ModelSchemaField) BufferFor(model QueryModel, optional bool) *SelectBuffer {
	modelValue := valueOf(model)
	modelField := modelValue.FieldByName(s.Name)
	selectBuffer := &SelectBuffer{}
	if IsField(s.Type) {
		selectBuffer.Value = ""
		selectBuffer.Ptr = &selectBuffer.Value
		// we make the pointer point to the value
		selectBuffer.Field = modelField.Interface().(Field)
	} else {
		selectBuffer.Ptr = modelField.Addr().Interface()
	}

	if optional {
		selectBuffer.Ptr = &Nullable{selectBuffer.Ptr}
	}

	return selectBuffer
}

func (s *ModelSchema) SelectBuffers(model QueryModel) []*SelectBuffer {

	selectBuffers := make([]*SelectBuffer, 0)
	modelValue := valueOf(model)
	for _, field := range s.Fields {
		if field.Optional && !s.IncludeOptional {
			continue
		}
		selectBuffer := field.BufferFor(model, s.Optional)
		selectBuffers = append(selectBuffers, selectBuffer)
	}
	for _, relatedModelSchema := range s.RelatedModelSchemas {
		modelField := modelValue.FieldByName(relatedModelSchema.Field)
		relatedModel := relatedModelSchema.ModelSchema.New(model.Database())
		modelField.Set(reflect.ValueOf(relatedModel))
		selectBuffers = append(selectBuffers, relatedModelSchema.ModelSchema.SelectBuffers(relatedModel)...)
	}
	return selectBuffers
}

func (s *ModelSchema) CheckForNull(model QueryModel) {
	modelValue := valueOf(model)

	// we go through all the related fields
	for _, relatedModelSchema := range s.RelatedModelSchemas {

		// we retrieve the foreign key field
		fkField := modelValue.FieldByName(relatedModelSchema.FkField)
		// and we retrieve the foreign model value
		modelField := modelValue.FieldByName(relatedModelSchema.Field)

		// if the foreign key isn't a pointer it can't be nil...
		if fkField.Type().Kind() != reflect.Ptr {
			continue
		}

		// if the foreign key is nil, we set the foreign model to nil as well...
		if fkField.IsNil() {
			modelField.Set(reflect.Zero(modelField.Type()))
			continue
		}

		// otherwise we recursively check the model itself...
		relatedModel := modelField.Interface().(QueryModel)
		relatedModelSchema.ModelSchema.CheckForNull(relatedModel)

	}
}

// https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6f
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func IsField(typ reflect.Type) bool {
	t := reflect.TypeOf((*Field)(nil)).Elem()
	if typ.Implements(t) {
		return true
	}
	return false
}

func IsUpdateField(typ reflect.Type) bool {
	t := reflect.TypeOf((*UpdateField)(nil)).Elem()
	if typ.Implements(t) {
		return true
	}
	return false
}

func extractField(fieldType reflect.Type, fieldValue reflect.Value, isPtr bool) *ModelSchemaField {
	switch fieldType.Kind() {
	case reflect.Struct:
		switch fieldType {
		case reflect.TypeOf(time.Time{}):
			break
		default:
			// this is not a supported struct
			return nil
		}
	case reflect.Slice:
		isPtr = true
	case reflect.Array:
		isPtr = true
	case reflect.String:
	case reflect.Int:
	case reflect.Bool:
	case reflect.Uint:
	case reflect.Int64:
	case reflect.Float64:
	case reflect.Ptr:
		// this might be a pointer to a relevant struct, so we check the type
		// of the pointed to value
		if IsField(fieldType) {
			isPtr = true
			break
		}
		return extractField(fieldType.Elem(), fieldValue, true)
	default:
		return nil
	}
	var value reflect.Value
	if isPtr {
		value = fieldValue
	} else {
		value = fieldValue.Addr()
	}
	return &ModelSchemaField{
		Value: value,
		Type:  fieldType,
	}
}

func HasTag(tags []Tag, key string) bool {
	for _, t := range tags {
		if t.Name == key {
			return true
		}
	}
	return false
}

func GetTag(tags []Tag, key string) (Tag, bool) {
	for _, t := range tags {
		if t.Name == key {
			return t, true
		}
	}
	return Tag{}, false
}

func extractFields(structValue reflect.Value) []*ModelSchemaField {
	fields := make([]*ModelSchemaField, 0)
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := structValue.Field(i)
		name := ToSnakeCase(fieldType.Name)
		tags := ExtractTags(fieldType)
		if HasTag(tags, "ignore") || !fieldValue.CanInterface() {
			continue
		}
		// checks if the field implements the query model interface
		if fieldType.Anonymous && fieldType.Type.Kind() == reflect.Struct {
			fields = append(fields, extractFields(fieldValue)...)
			continue
		}
		field := extractField(fieldType.Type, fieldValue, false)
		if field == nil {
			continue
		}
		field.Name = fieldType.Name
		field.Column = name
		// we allow overwriting of the column name
		if HasTag(tags, "col") {
			tag, _ := GetTag(tags, "col")
			field.Column = tag.Value
		}
		if HasTag(tags, "optional") {
			field.Optional = true
		}
		field.Tags = tags
		fields = append(fields, field)
	}
	return fields
}

func addRelated(relatedModelSchemas []*RelatedModelSchema, tableName string) ([]string, []string, error) {

	selectNames := make([]string, 0)
	joins := make([]string, 0)

	for _, relatedModelSchema := range relatedModelSchemas {
		// we initialize a model instance
		relatedTableName := relatedModelSchema.ModelSchema.TableName
		var relatedPkField *ModelSchemaField
		for _, field := range relatedModelSchema.ModelSchema.Fields {
			if HasTag(field.Tags, "pk") {
				relatedPkField = field
				break
			}
		}
		if relatedPkField == nil {
			return nil, nil, fmt.Errorf("missing primary key")
		}

		join := "INNER JOIN"

		if relatedModelSchema.Optional {
			// this field is optional
			join = "LEFT JOIN"
		}

		qualifiedName := fmt.Sprintf("%s_%s_%s", tableName, relatedModelSchema.Column, relatedTableName)
		condition := fmt.Sprintf("%s \"%s\" %s ON %s.%s = %s.%s AND %s.deleted_at IS NULL", join, relatedTableName, qualifiedName, qualifiedName, relatedPkField.Column, tableName, relatedModelSchema.Column, qualifiedName)
		joins = append(joins, condition)
		for _, field := range relatedModelSchema.ModelSchema.Fields {
			// we always skip optional fields for related models
			if field.Optional {
				continue
			}
			selectNames = append(selectNames, fmt.Sprintf("%s.%s", qualifiedName, field.Column))
		}
		var relatedJoins, relatedSelectNames []string
		var err error
		if relatedJoins, relatedSelectNames, err = addRelated(relatedModelSchema.ModelSchema.RelatedModelSchemas, qualifiedName); err != nil {
			return nil, nil, err
		}
		joins = append(joins, relatedJoins...)
		selectNames = append(selectNames, relatedSelectNames...)
	}
	return joins, selectNames, nil
}

func LoadOne(model QueryModel, queries map[string]interface{}) error {
	if _, err := Load(model, queries, true); err != nil {
		return err
	} else {
		return nil
	}
}

func LoadMany(model QueryModel, queries map[string]interface{}) ([]QueryModel, error) {
	return Load(model, queries, false)
}

type QueryStmt struct {
	Model       QueryModel
	ModelSchema *ModelSchema
	Query       string
}

func (q *QueryStmt) Execute(args ...any) ([]QueryModel, error) {

	models := make([]QueryModel, 0)

	if rows, err := q.Model.Database()().Query(q.Query, args...); err != nil {
		return nil, err
	} else {
		defer rows.Close()

		for rows.Next() {
			nextModel := q.ModelSchema.New(q.Model.Database())
			selectBuffers := q.ModelSchema.SelectBuffers(nextModel)
			buffers := make([]interface{}, len(selectBuffers))

			for i, selectBuffer := range selectBuffers {
				buffers[i] = selectBuffer.Ptr
			}

			if err := rows.Scan(buffers...); err != nil {
				return nil, err
			}

			// we initialize all buffers
			for _, selectBuffer := range selectBuffers {
				if err := selectBuffer.Init(); err != nil {
					return nil, err
				}
			}

			q.ModelSchema.CheckForNull(nextModel)
			models = append(models, nextModel)
		}
	}

	return models, nil
}

type LoadStmt struct {
	Single          bool
	Model           QueryModel
	ModelSchema     *ModelSchema
	Conditions      []string
	ConditionValues []interface{}
	SelectNames     []string
	Joins           []string
	OrderBy         string
	Limit           string
	Offset          string
}

func (l *LoadStmt) AddValue(value interface{}) int {
	l.ConditionValues = append(l.ConditionValues, value)
	return len(l.ConditionValues)
}

func (l *LoadStmt) Execute() ([]QueryModel, error) {

	models := make([]QueryModel, 0)

	if rows, err := l.Model.Database()().Query(l.Statement(), l.ConditionValues...); err != nil {
		return nil, err
	} else {
		defer rows.Close()

		for rows.Next() {
			var nextModel QueryModel
			if l.Single {
				nextModel = l.Model
			} else {
				nextModel = l.ModelSchema.New(l.Model.Database())
			}

			selectBuffers := l.ModelSchema.SelectBuffers(nextModel)
			buffers := make([]interface{}, len(selectBuffers))

			for i, selectBuffer := range selectBuffers {
				buffers[i] = selectBuffer.Ptr
			}

			if err := rows.Scan(buffers...); err != nil {
				return nil, err
			}

			// we initialize all buffers
			for _, selectBuffer := range selectBuffers {
				if err := selectBuffer.Init(); err != nil {
					return nil, err
				}
			}

			l.ModelSchema.CheckForNull(nextModel)

			models = append(models, nextModel)
			if l.Single {
				break
			}
		}
	}

	return models, nil
}

func (l *LoadStmt) AddJoin(join string) {
	l.Joins = append(l.Joins, join)
}

func (l *LoadStmt) AddCondition(condition string) {
	l.Conditions = append(l.Conditions, condition)
}

func (l *LoadStmt) Statement() string {
	return fmt.Sprintf(`
	SELECT
		%s
	FROM
		"%s"
	%s
	WHERE
		%s
	%s
	%s
	%s
		`, strings.Join(l.SelectNames, ", "), l.Model.TableName(), strings.Join(l.Joins, "\n"), strings.Join(l.Conditions, " AND "), l.OrderBy, l.Limit, l.Offset)
}

type Condition interface {
	Value() []interface{}
	Condition(string, int) string
}

type Comparison struct {
	Op      string
	Operand interface{}
}

func (i Comparison) Value() []interface{} {
	return []interface{}{i.Operand}
}

func (i Comparison) Condition(key string, parameter int) string {
	return fmt.Sprintf("%s %s $%d", key, i.Op, parameter)
}

type In struct {
	Values []interface{}
}

type IsNil struct {
}

type IsNotNil struct {
}

func (i IsNil) Value() []interface{} {
	return nil
}

func (i IsNil) Condition(key string, parameter int) string {
	return fmt.Sprintf("%s IS NULL", key)
}

func (i IsNotNil) Value() []interface{} {
	return nil
}

func (i IsNotNil) Condition(key string, parameter int) string {
	return fmt.Sprintf("%s IS NOT NULL", key)
}

func (i In) Value() []interface{} {
	return i.Values
}

func (i In) Condition(key string, parameter int) string {
	keys := make([]string, len(i.Values))
	for j := 0; j < len(i.Values); j++ {
		keys[j] = fmt.Sprintf("$%d", parameter+j)
	}
	// we return a FALSE statement if no values are given in the IN clause, as
	// otherwise Postgres will throw an error
	if len(i.Values) == 0 {
		return fmt.Sprintf("FALSE")
	}
	return fmt.Sprintf("%s IN (%s)", key, strings.Join(keys, ", "))
}

func getConditions(tableName string, queries map[string]interface{}, offset int) ([]string, []interface{}) {
	conditions := make([]string, 0)
	conditionValues := make([]interface{}, 0)

	conditions = append(conditions, fmt.Sprintf("\"%s\".deleted_at IS NULL", tableName))

	for key, value := range queries {
		// if the key is qualified we do not qualify it again
		qualifiedKey := fmt.Sprintf("\"%s\".%s", tableName, key)
		if strings.Contains(key, ".") {
			qualifiedKey = key
		}
		if condition, ok := value.(Condition); ok {
			cv := condition.Condition(qualifiedKey, len(conditionValues)+1+offset)
			conditions = append(conditions, cv)
			cvv := condition.Value()
			if cvv != nil {
				conditionValues = append(conditionValues, condition.Value()...)
			}
		} else {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", qualifiedKey, len(conditionValues)+1+offset))
			conditionValues = append(conditionValues, value)
		}
	}
	return conditions, conditionValues
}

func DeleteMany(model QueryModel, queries map[string]interface{}) error {
	stmt := GetDeleteStmt(model, queries)
	return stmt.Execute()
}

type UpdateStmt struct {
	Model           QueryModel
	Conditions      []string
	ConditionValues []interface{}
	Keys            []string
	Values          []interface{}
}

type Expression struct {
	Expression string
}

func Update(model QueryModel, queries map[string]interface{}, data map[string]interface{}) error {
	stmt := GetUpdateStmt(model, queries, data)
	return stmt.Execute()
}

func GetUpdateStmt(model QueryModel, queries map[string]interface{}, data map[string]interface{}) *UpdateStmt {
	keys := make([]string, 0)
	values := make([]interface{}, 0)

	for key, value := range data {
		if expressionValue, ok := value.(Expression); ok {
			// this is dangerous...
			keys = append(keys, fmt.Sprintf("%s = %s", key, expressionValue.Expression))
		} else {
			keys = append(keys, fmt.Sprintf("%s = $%d", key, len(values)+1))
			values = append(values, value)
		}
	}

	tableName := model.TableName()
	conditions, conditionValues := getConditions(tableName, queries, len(values))

	return &UpdateStmt{
		Model:           model,
		Keys:            keys,
		Values:          values,
		Conditions:      conditions,
		ConditionValues: conditionValues,
	}

}

func (u *UpdateStmt) Statement() string {
	return fmt.Sprintf(`
	UPDATE
		"%s"
	SET
		%s
	WHERE
		%s`, u.Model.TableName(), strings.Join(u.Keys, ", "), strings.Join(u.Conditions, " AND "))
}

func (d *UpdateStmt) Execute() error {
	if _, err := d.Model.Database()().Exec(d.Statement(), append(d.Values, d.ConditionValues...)...); err != nil {
		return err
	} else {
		return nil
	}
}

type DeleteStmt struct {
	Model           QueryModel
	Conditions      []string
	ConditionValues []interface{}
}

func (d *DeleteStmt) Statement() string {
	return fmt.Sprintf(`
	DELETE FROM
		"%s"
	WHERE
		%s`, d.Model.TableName(), strings.Join(d.Conditions, " AND "))
}

func (d *DeleteStmt) Execute() error {
	if _, err := d.Model.Database()().Exec(d.Statement(), d.ConditionValues...); err != nil {
		return err
	} else {
		return nil
	}
}

func GetDeleteStmt(model QueryModel, queries map[string]interface{}) *DeleteStmt {
	tableName := model.TableName()
	conditions, conditionValues := getConditions(tableName, queries, 0)

	return &DeleteStmt{
		Model:           model,
		Conditions:      conditions,
		ConditionValues: conditionValues,
	}

}

func GetQueryStmt(model QueryModel, query string) *QueryStmt {
	return &QueryStmt{
		Model:       model,
		ModelSchema: InferModelSchema(model),
		Query:       query,
	}
}

func GetLoadStmt(model QueryModel, queries map[string]interface{}, single bool) (*LoadStmt, error) {

	var joins, relatedSelectNames []string
	var err error

	modelSchema := InferModelSchema(model)

	if single {
		modelSchema.IncludeOptional = true
	}

	tableName := model.TableName()
	selectNames := make([]string, 0)

	conditions, conditionValues := getConditions(tableName, queries, 0)

	for _, field := range modelSchema.Fields {
		if field.Optional && !modelSchema.IncludeOptional {
			continue
		}
		selectNames = append(selectNames, fmt.Sprintf("\"%s\".%s", tableName, field.Column))
	}

	if joins, relatedSelectNames, err = addRelated(modelSchema.RelatedModelSchemas, tableName); err != nil {
		return nil, err
	}

	selectNames = append(selectNames, relatedSelectNames...)

	return &LoadStmt{
		Single:          single,
		Model:           model,
		ModelSchema:     modelSchema,
		SelectNames:     selectNames,
		Joins:           joins,
		ConditionValues: conditionValues,
		Conditions:      conditions,
	}, nil
}

func Load(model QueryModel, queries map[string]interface{}, single bool) ([]QueryModel, error) {

	stmt, err := GetLoadStmt(model, queries, single)

	if err != nil {
		return nil, err
	}

	return LoadWithStmt(stmt)
}

func LoadWithStmt(stmt *LoadStmt) ([]QueryModel, error) {
	models, err := stmt.Execute()

	if err != nil {
		return nil, err
	}

	if stmt.Single && len(models) == 0 {
		return nil, NotFound
	}

	return models, nil

}

func Refresh(model QueryModel) error {
	if err := model.Init(); err != nil {
		return err
	}
	if dbModel, ok := model.(*DBModel); !ok {
		return fmt.Errorf("refreshing only supported for DB models")
	} else {
		return LoadOne(model, map[string]interface{}{"id": dbModel.ID})
	}
}

func Delete(model QueryModel) error {
	if dbModel, ok := model.(*DBModel); !ok {
		return fmt.Errorf("deleting only supported for DB models")
	} else {
		deleteQuery := fmt.Sprintf(`
		UPDATE
			"%s"
		SET
			deleted_at = current_timestamp
		WHERE id = $1
		`, model.TableName())

		_, err := model.Database()().Exec(deleteQuery, dbModel.ID)

		if err != nil {
			return err
		}

	}
	return nil
}

func IsNull(value reflect.Value) bool {
	return value.IsNil() || unpointValue(value).IsZero()
}

func Save(model QueryModel) error {

	modelSchema := InferModelSchema(model)

	insertNames := make([]string, 0)
	insertPlaceholders := make([]string, 0)
	insertValues := make([]interface{}, 0)
	updateNames := make([]string, 0)
	conflictNames := make([]string, 0)
	returnNames := make([]string, 0)
	returnBuffers := make([]*SelectBuffer, 0)
	returnValues := make([]interface{}, 0)

	tableName := model.TableName()
	p := 1
	for _, field := range modelSchema.Fields {

		// we skip optional fields if they are null
		if field.Optional && IsNull(field.Value) {
			continue
		}

		if HasTag(field.Tags, "onConflict") || (HasTag(field.Tags, "pk") && !HasTag(field.Tags, "noOnConflict")) {
			conflictNames = append(conflictNames, field.Column)
		}
		// if the field has an "autogen" tag and is zero, we generate a value
		if HasTag(field.Tags, "autogen") && IsNull(field.Value) && IsField(field.Type) {
			if err := field.Value.Interface().(Field).Generate(); err != nil {
				return err
			}
		}
		if HasTag(field.Tags, "update") {
			if IsUpdateField(field.Type) {
				if err := field.Value.Interface().(UpdateField).Update(); err != nil {
					return err
				}
			} else if updateModel, ok := model.(UpdateModel); ok {
				if err := updateModel.UpdateField(field.Name); err != nil {
					return err
				}
			}
		}
		// if this field is autogenerated (either on the DB-side or client-side)
		// we return its value from the SELECT statement
		if HasTag(field.Tags, "auto") || HasTag(field.Tags, "autogen") || HasTag(field.Tags, "readAfterWrite") {
			returnNames = append(returnNames, field.Column)
			returnBuffer := field.BufferFor(model, false)
			returnValues = append(returnValues, returnBuffer.Ptr)
			returnBuffers = append(returnBuffers, returnBuffer)
		}
		// we provide the value unless it is autogenerated on the DB-side
		if !(HasTag(field.Tags, "auto") && IsNull(field.Value)) {
			insertNames = append(insertNames, field.Column)
			v := field.Value.Interface()
			if vf, ok := v.(Field); ok {
				gv := vf.Get()
				if gv == nil {
					gv = nil
				} else {
					vo := reflect.ValueOf(gv)

					switch vo.Kind() {
					case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
						if vo.IsNil() {
							gv = nil
						}
					}
				}
				insertValues = append(insertValues, gv)
			} else {
				insertValues = append(insertValues, v)
			}
			insertPlaceholders = append(insertPlaceholders, fmt.Sprintf("$%d", p))
			// if the field is autogenerated we only update it
			if HasTag(field.Tags, "auto") || HasTag(field.Tags, "autogen") {
				// if this is autogenerated we only update it if it's not yet set
				updateNames = append(updateNames, fmt.Sprintf("%s = COALESCE(\"%s\".%s, $%d)", field.Column, tableName, field.Column, p))
			} else {
				updateNames = append(updateNames, fmt.Sprintf("%s = $%d", field.Column, p))
			}
			p = p + 1
		}
	}

	returning := ""

	if len(returnNames) > 0 {
		returning = fmt.Sprintf(`
		RETURNING
		%s
		`, strings.Join(returnNames, ", "))
	}

	query := fmt.Sprintf(`
	INSERT INTO
		"%s"
		(%s)
	VALUES
		(%s)
	ON CONFLICT (%s) WHERE deleted_at IS NULL DO UPDATE SET %s
	%s
	`, tableName,
		strings.Join(insertNames, ", "),
		strings.Join(insertPlaceholders, ", "),
		strings.Join(conflictNames, ", "),
		strings.Join(updateNames, ", "),
		returning,
	)

	if rows, err := model.Database()().Query(query, insertValues...); err != nil {
		return err
	} else {

		defer rows.Close()

		if len(returnValues) > 0 {

			found := rows.Next()

			if !found {
				return fmt.Errorf("cannot insert")
			}

			if err := rows.Scan(returnValues...); err != nil {
				return err
			}

			for _, buffer := range returnBuffers {
				if err := buffer.Init(); err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func ExtractTags(field reflect.StructField) []Tag {
	tags := make([]Tag, 0)
	if dbValue, ok := field.Tag.Lookup("db"); ok {
		strTags := strings.Split(dbValue, ",")
		for _, tag := range strTags {
			kv := strings.Split(dbValue, ":")
			if len(kv) == 1 {
				tags = append(tags, Tag{
					Name:  tag,
					Value: "",
					Flag:  true,
				})
			} else {
				tags = append(tags, Tag{
					Name:  kv[0],
					Value: kv[1],
					Flag:  false,
				})
			}
		}
	}
	return tags
}

func InferModelSchema(model QueryModel) *ModelSchema {
	return &ModelSchema{
		TableName:           model.TableName(),
		Type:                typeOf(model),
		Fields:              extractFields(valueOf(model)),
		RelatedModelSchemas: InferRelatedModelSchemas(model),
	}
}

func New(modelType reflect.Type) QueryModel {
	model := reflect.New(modelType).Interface().(QueryModel)
	Init(model, nil)
	return model
}

func InferRelatedModelSchemas(model QueryModel) []*RelatedModelSchema {
	relatedModelSchemas := make([]*RelatedModelSchema, 0)
	modelValue := valueOf(model)
	modelType := modelValue.Type()
	for i := 0; i < modelValue.NumField(); i++ {
		fieldType := modelType.Field(i)

		tags := ExtractTags(fieldType)
		if tag, ok := GetTag(tags, "fk"); ok {

			optional := false

			for j := 0; j < modelValue.NumField(); j++ {
				fkFieldType := modelType.Field(j)

				if fkFieldType.Name != tag.Value {
					continue
				}

				if fkFieldType.Type.Kind() == reflect.Ptr {
					optional = true
				}

				break

			}

			modelSchema := InferModelSchema(New(unpointType(fieldType.Type)))

			modelSchema.Optional = optional

			relatedModelSchema := &RelatedModelSchema{
				Column:      ToSnakeCase(tag.Value),
				FkField:     tag.Value,
				Optional:    optional,
				Field:       fieldType.Name,
				ModelSchema: modelSchema,
			}
			relatedModelSchemas = append(relatedModelSchemas, relatedModelSchema)
		}
	}
	return relatedModelSchemas
}

func unpointValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return reflect.Indirect(value)
	}
	return value
}

func unpointType(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		return typ.Elem()
	}
	return typ
}

func valueOf(model QueryModel) reflect.Value {
	return unpointValue(reflect.ValueOf(model))
}

func typeOf(model QueryModel) reflect.Type {
	return unpointType(reflect.TypeOf(model))
}

func InitType[T QueryModel](db func() DB) T {
	var obj T
	obj = reflect.New(reflect.TypeOf(obj).Elem()).Interface().(T)
	Init(obj, db)
	return obj
}

func Init[T QueryModel](model T, db func() DB) T {
	modelValue := valueOf(model)
	if model.TableName() == "" {
		model.SetTableName(ToSnakeCase(modelValue.Type().Name()))
	}
	model.SetDatabase(db)
	inferredModelSchema := InferModelSchema(model)
	for _, field := range inferredModelSchema.Fields {
		// if this is a specific field, we initialize it
		if IsField(field.Type) {
			if IsNull(field.Value) {
				field.Value.Set(reflect.New(unpointType(field.Type)))
				field.Value.Interface().(Field).Init()
			}
		}
	}
	return model
}

// assigns the model to the foreign key field
func (r *RelatedModelSchema) Assign(model, relatedModel QueryModel) error {
	value := reflect.ValueOf(model)
	relatedValue := reflect.ValueOf(relatedModel)
	field := value.FieldByName(r.Field)
	if !field.CanSet() {
		return fmt.Errorf("field does not exist")
	}
	field.Set(relatedValue)
	return nil
}
