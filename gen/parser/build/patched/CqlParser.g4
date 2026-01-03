/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 by Domagoj Kovačević
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
 * associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute,
 * sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or
 * substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
 * NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Project : cql-parser; an ANTLR4 grammar for Apache Cassandra CQL  https://github.com/kdcro101cql-parser
 */

// $antlr-format alignTrailingComments true, columnLimit 150, minEmptyLines 1, maxEmptyLinesToKeep 1, reflowComments false, useTab false
// $antlr-format allowShortRulesOnASingleLine false, allowShortBlocksOnASingleLine true, alignSemicolons hanging, alignColons hanging

parser grammar CqlParser;

options
   {
    tokenVocab = CqlLexer;
}

root
    : cqls? MINUSMINUS? EOF
    ;

cqls
    : (cql MINUSMINUS? statementSeparator | empty_)* (
        cql (MINUSMINUS? statementSeparator)?
        | empty_
    )
    ;

statementSeparator
    : SEMI
    ;

empty_
    : statementSeparator
    ;

cql
    : alterKeyspace
    | alterMaterializedView
    | alterRole
    | alterTable
    | alterType
    | alterUser
    | applyBatch
    | createAggregate
    | createFunction
    | createIndex
    | createKeyspace
    | createMaterializedView
    | createRole
    | createTable
    | createTrigger
    | createType
    | createUser
    | delete_
    | dropAggregate
    | dropFunction
    | dropIndex
    | dropKeyspace
    | dropMaterializedView
    | dropRole
    | dropTable
    | dropTrigger
    | dropType
    | dropUser
    | grant
    | insert
    | listPermissions
    | listRoles
    | revoke
    | select_
    | truncate
    | update
    | use_
    
    | pruneMaterializedView  // ScyllaDB extension
    | batch               // Multi-statement batch
    | describeStatement   // DESCRIBE/DESC
    | createServiceLevel  // Service Level management
    | alterServiceLevel
    | dropServiceLevel
    | attachServiceLevel
    | detachServiceLevel
    | listServiceLevel
    
    | listUsers           // LIST USERS
    ;

revoke
    : kwRevoke priviledge kwOn resource kwFrom role
    ;

listRoles
    : kwList kwAll? kwRoles (kwOf role)? kwNorecursive?
    ;

// LIST USERS statement
listUsers
    : kwList kwUsers
    ;

listPermissions
    : kwList priviledge (kwOn resource)? (kwOf role)?
    ;

grant
    : kwGrant priviledge kwOn resource kwTo role
    ;

priviledge
    : (kwAll | kwAllPermissions)
    | kwAlter
    | kwAuthorize
    | kwDescibe
    | kwExecute
    | kwCreate
    | kwDrop
    | kwModify
    | kwSelect
    
    | kwVectorSearchIndexing   // Vector search indexing permission
    ;

resource
    : kwAll kwFunctions
    | kwAll kwFunctions kwIn kwKeyspace keyspace
    | kwFunction (keyspace DOT)? function_
    | kwAll kwKeyspaces
    | kwKeyspace keyspace
    | (kwTable)? (keyspace DOT)? table
    | kwAll kwRoles
    | kwRole role
    ;

createUser
    : kwCreate kwUser ifNotExist? user kwWith kwPassword stringLiteral (
        kwSuperuser
        | kwNosuperuser
    )?
    ;

createRole
    : kwCreate kwRole ifNotExist? role roleWith?
    ;

createType
    : kwCreate kwType ifNotExist? (keyspace DOT)? type_ syntaxBracketLr typeMemberColumnList syntaxBracketRr
    ;

typeMemberColumnList
    : column dataType (syntaxComma column dataType)*
    ;

createTrigger
    : kwCreate kwTrigger ifNotExist? (keyspace DOT)? trigger kwUsing triggerClass
    ;

createMaterializedView
    : kwCreate kwMaterialized kwView ifNotExist? (keyspace DOT)? materializedView kwAs kwSelect selectElements fromSpec mvWhereSpec? primaryKeyElement (kwWith materializedViewOptions)?
    ;

// Materialized view WHERE clause with IS NOT NULL support
mvWhereSpec
    : kwWhere mvWhereClause (kwAnd mvWhereClause)*
    ;

mvWhereClause
    : columnRef kwIs kwNot kwNull
    | columnRef (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) constant
    ;

materializedViewWhere
    : kwWhere columnNotNullList (kwAnd relationElements)?
    ;

columnNotNullList
    : columnNotNull (kwAnd columnNotNull)*
    ;

columnNotNull
    : column kwIs kwNot kwNull
    ;

materializedViewOptions
    : tableOptions
    | tableOptions kwAnd clusteringOrder
    | clusteringOrder
    ;

// CREATE MATERIALIZED VIEW [IF NOT EXISTS] [keyspace_name.] view_name
// AS SELECT column_list
// FROM [keyspace_name.] base_table_name
// WHERE column_name IS NOT NULL [AND column_name IS NOT NULL ...]
//       [AND relation...]
// PRIMARY KEY ( column_list )
// [WITH [table_properties]
//       [AND CLUSTERING ORDER BY (cluster_column_name order_option )]]
createKeyspace
    : kwCreate kwKeyspace ifNotExist? keyspace kwWith kwReplication OPERATOR_EQ syntaxBracketLc replicationList syntaxBracketRc (kwAnd durableWrites)? (kwAnd tabletsSpec)?
    ;

createFunction
    : kwCreate orReplace? kwFunction ifNotExist? (keyspace DOT)? function_ syntaxBracketLr paramList? syntaxBracketRr returnMode kwReturns dataType
        kwLanguage language kwAs codeBlock
    ;

codeBlock
    : CODE_BLOCK
    | STRING_LITERAL
    ;

paramList
    : param (syntaxComma param)*
    ;

returnMode
    : (kwCalled | kwReturns kwNull) kwOn kwNull kwInput
    ;

createAggregate
    : kwCreate orReplace? kwAggregate ifNotExist? (keyspace DOT)? aggregate syntaxBracketLr dataType syntaxBracketRr
      kwSfunc function_ kwStype dataType
      (kwReducefunc function_)?
      (kwFinalfunc function_)?
      (kwInitcond initCondDefinition)?
    ;

// paramList
// :
initCondDefinition
    : constant
    | initCondList
    | initCondListNested
    | initCondHash
    ;

initCondHash
    : syntaxBracketLc initCondHashItem (syntaxComma initCondHashItem)* syntaxBracketRc
    ;

initCondHashItem
    : hashKey COLON initCondDefinition
    ;

initCondListNested
    : syntaxBracketLr initCondList (syntaxComma constant | initCondList)* syntaxBracketRr
    ;

initCondList
    : syntaxBracketLr constant (syntaxComma constant)* syntaxBracketRr
    ;

orReplace
    : kwOr kwReplace
    ;

alterUser
    : kwAlter kwUser user kwWith userPassword userSuperUser?
    ;

userPassword
    : kwPassword stringLiteral
    ;

userSuperUser
    : kwSuperuser
    | kwNosuperuser
    ;

alterType
    : kwAlter kwType (keyspace DOT)? type_ alterTypeOperation
    ;

alterTypeOperation
    : alterTypeAlterType
    | alterTypeAdd
    | alterTypeRename
    ;

alterTypeRename
    : kwRename alterTypeRenameList
    ;

alterTypeRenameList
    : alterTypeRenameItem (kwAnd alterTypeRenameItem)*
    ;

alterTypeRenameItem
    : column kwTo column
    ;

alterTypeAdd
    : kwAdd column dataType (syntaxComma column dataType)*
    ;

alterTypeAlterType
    : kwAlter column kwType dataType
    ;

alterTable
    : kwAlter kwTable (keyspace DOT)? table alterTableOperation
    ;

alterTableOperation
    : alterTableAdd
    | alterTableDropColumns
    | alterTableDropCompactStorage
    | alterTableRename
    | alterTableWith
    ;

alterTableWith
    : kwWith tableOptions
    ;

alterTableRename
    : kwRename column kwTo column
    ;

alterTableDropCompactStorage
    : kwDrop kwCompact kwStorage
    ;

alterTableDropColumns
    : kwDrop alterTableDropColumnList
    ;

alterTableDropColumnList
    : column (syntaxComma column)*
    ;

alterTableAdd
    : kwAdd column dataType staticColumn?
    | kwAdd syntaxBracketLr columnDefinition (syntaxComma columnDefinition)* syntaxBracketRr
    ;

alterTableColumnDefinition
    : column dataType (syntaxComma column dataType)*
    ;

alterRole
    : kwAlter kwRole role roleWith?
    ;

roleWith
    : kwWith (roleWithOptions (kwAnd roleWithOptions)*)
    ;

roleWithOptions
    : kwPassword OPERATOR_EQ stringLiteral
    | kwHashed kwPassword OPERATOR_EQ stringLiteral
    | kwLogin OPERATOR_EQ booleanLiteral
    | kwSuperuser OPERATOR_EQ booleanLiteral
    | kwOptions OPERATOR_EQ optionHash
    | kwNologin OPERATOR_EQ booleanLiteral
    ;

alterMaterializedView
    : kwAlter kwMaterialized kwView (keyspace DOT)? materializedView (kwWith tableOptions)?
    ;

dropUser
    : kwDrop kwUser ifExist? user
    ;

dropType
    : kwDrop kwType ifExist? (keyspace DOT)? type_
    ;

dropMaterializedView
    : kwDrop kwMaterialized kwView ifExist? (keyspace DOT)? materializedView
    ;

dropAggregate
    : kwDrop kwAggregate ifExist? (keyspace DOT)? aggregate (syntaxBracketLr dataType syntaxBracketRr)?
    ;

dropFunction
    : kwDrop kwFunction ifExist? (keyspace DOT)? function_
    ;

dropTrigger
    : kwDrop kwTrigger ifExist? trigger kwOn (keyspace DOT)? table
    ;

dropRole
    : kwDrop kwRole ifExist? role
    ;

dropTable
    : kwDrop kwTable ifExist? (keyspace DOT)? table
    ;

dropKeyspace
    : kwDrop kwKeyspace ifExist? keyspace
    ;

dropIndex
    : kwDrop kwIndex ifExist? (keyspace DOT)? indexName
    ;

createTable
    : kwCreate kwTable ifNotExist? (keyspace DOT)? table syntaxBracketLr columnDefinitionList syntaxBracketRr withElement?
    ;

withElement
    : kwWith tableOptions
    ;

tableOptions
    : kwCompact kwStorage (kwAnd tableOptions)?
    | clusteringOrder (kwAnd tableOptions)?
    | tableOptionItem (kwAnd tableOptionItem)*
    ;

clusteringOrder
    : kwClustering kwOrder kwBy syntaxBracketLr (column orderDirection?) (syntaxComma column orderDirection?)* syntaxBracketRr
    ;

tableOptionItem
    : tableOptionName OPERATOR_EQ tableOptionValue
    | tableOptionName OPERATOR_EQ optionHash
    ;

tableOptionName
    : OBJECT_NAME
    ;

tableOptionValue
    : stringLiteral
    | floatLiteral
    ;

optionHash
    : syntaxBracketLc optionHashItem (syntaxComma optionHashItem)* syntaxBracketRc
    ;

optionHashItem
    : optionHashKey COLON optionHashValue
    ;

optionHashKey
    : stringLiteral
    ;

optionHashValue
    : stringLiteral
    | floatLiteral
    | booleanLiteral
    | decimalLiteral
    ;

columnDefinitionList
    : (columnDefinition) (syntaxComma columnDefinition)* (syntaxComma primaryKeyElement)?
    ;

//
columnDefinition
    : column dataType staticColumn? primaryKeyColumn?
    ;

//
primaryKeyColumn
    : kwPrimary kwKey
    ;

// ScyllaDB/Cassandra: STATIC column modifier
staticColumn
    : kwStatic
    ;

primaryKeyElement
    : kwPrimary kwKey syntaxBracketLr primaryKeyDefinition syntaxBracketRr
    ;

primaryKeyDefinition
    : singlePrimaryKey
    | compoundKey
    | compositeKey
    ;

singlePrimaryKey
    : column
    ;

compoundKey
    : partitionKey (syntaxComma clusteringKeyList)
    ;

compositeKey
    : syntaxBracketLr partitionKeyList syntaxBracketRr (syntaxComma clusteringKeyList)
    ;

partitionKeyList
    : (partitionKey) (syntaxComma partitionKey)*
    ;

clusteringKeyList
    : (clusteringKey) (syntaxComma clusteringKey)*
    ;

partitionKey
    : column
    ;

clusteringKey
    : column
    ;

applyBatch
    : kwApply kwBatch
    ;

// Multi-statement batch: BEGIN BATCH stmt1 stmt2 ... APPLY BATCH
batch
    : kwBegin batchType? kwBatch usingTimestampSpec? batchStatementList? kwApply kwBatch
    ;

batchStatementList
    : batchStatement+
    ;

batchStatement
    : batchInsert
    | batchUpdate
    | batchDelete
    ;

batchInsert
    : kwInsert kwInto (keyspace DOT)? table insertColumnSpec? insertValuesSpec ifNotExist? usingTtlTimestamp?
    ;

batchUpdate
    : kwUpdate (keyspace DOT)? table usingTtlTimestamp? kwSet assignments whereSpec (ifExist | ifSpec)?
    ;

batchDelete
    : kwDelete deleteColumnList? fromSpec usingTimestampSpec? whereSpec (ifExist | ifSpec)?
    ;

beginBatch
    : kwBegin batchType? kwBatch usingTimestampSpec?
    ;

batchType
    : kwLogged
    | kwUnlogged
    | kwCounter
    ;

alterKeyspace
    : kwAlter kwKeyspace keyspace kwWith kwReplication OPERATOR_EQ syntaxBracketLc replicationList syntaxBracketRc (
        kwAnd durableWrites
    )?
    ;

replicationList
    : (replicationListItem) (syntaxComma replicationListItem)*
    ;

replicationListItem
    : STRING_LITERAL COLON STRING_LITERAL
    | STRING_LITERAL COLON DECIMAL_LITERAL
    ;

durableWrites
    : kwDurableWrites OPERATOR_EQ booleanLiteral
    ;

// ScyllaDB tablets option
tabletsSpec
    : kwTablets OPERATOR_EQ syntaxBracketLc tabletsOptions syntaxBracketRc
    ;

tabletsOptions
    : tabletsOption (syntaxComma tabletsOption)*
    ;

tabletsOption
    : stringLiteral COLON (stringLiteral | booleanLiteral | decimalLiteral)
    ;

use_
    : kwUse keyspace
    ;

truncate
    : kwTruncate (kwTable)? (keyspace DOT)? table
    ;

createIndex
    : kwCreate kwCustom? kwIndex ifNotExist? OBJECT_NAME? kwOn (keyspace DOT)? table syntaxBracketLr indexColumnSpec syntaxBracketRr indexUsing?
    ;

// CREATE INDEX ... USING for custom index types
indexUsing
    : kwUsing stringLiteral indexOptions?
    ;

indexOptions
    : kwWith kwOptions OPERATOR_EQ optionHash
    ;

indexName
    : OBJECT_NAME
    | stringLiteral
    ;

indexColumnSpec
    : column
    | indexKeysSpec
    | indexEntriesSSpec
    | indexFullSpec
    ;

indexKeysSpec
    : kwKeys syntaxBracketLr OBJECT_NAME syntaxBracketRr
    ;

indexEntriesSSpec
    : kwEntries syntaxBracketLr OBJECT_NAME syntaxBracketRr
    ;

indexFullSpec
    : kwFull syntaxBracketLr OBJECT_NAME syntaxBracketRr
    ;

delete_
    : beginBatch? kwDelete deleteColumnList? fromSpec usingTimestampSpec? whereSpec (
        ifExist
        | ifSpec
    )?
    ;

deleteColumnList
    : (deleteColumnItem) (syntaxComma deleteColumnItem)*
    ;

deleteColumnItem
    : OBJECT_NAME
    | OBJECT_NAME LS_BRACKET (stringLiteral | decimalLiteral) RS_BRACKET
    ;

update
    : beginBatch? kwUpdate (keyspace DOT)? table usingTtlTimestamp? kwSet assignments whereSpec (
        ifExist
        | ifSpec
    )?
    ;

ifSpec
    : kwIf ifConditionList
    | kwIf kwNot kwExists
    ;

ifConditionList
    : (ifCondition) (kwAnd ifCondition)* (kwOr ifCondition)*
    ;

ifCondition
    : columnRef (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE | OPERATOR_NEQ) ifConditionValue
    | ifConditionValue (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE | OPERATOR_NEQ) columnRef
    | columnRef kwIn syntaxBracketLr (ifConditionValue (syntaxComma ifConditionValue)*)? syntaxBracketRr
    | columnRef syntaxBracketLs ifConditionValue syntaxBracketRs (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE | OPERATOR_NEQ) ifConditionValue
    | columnRef kwContains ifConditionValue
    | columnRef kwContains kwKey ifConditionValue
    | columnRef syntaxBracketLc ifConditionValue syntaxBracketRc (OPERATOR_EQ | OPERATOR_NEQ) ifConditionValue
    | columnRef kwLike ifConditionValue
    | columnRef
    | columnRef syntaxBracketLs ifConditionValue syntaxBracketRs
    ;

// IF condition values - support constants, collections, column references, functions
ifConditionValue
    : constant
    | OBJECT_NAME
    | functionCall
    | assignmentSet
    | assignmentList
    | assignmentMap
    ;

assignments
    : (assignmentElement) (syntaxComma assignmentElement)*
    ;

assignmentElement
    : columnRef OPERATOR_EQ (constant | assignmentMap | assignmentSet | assignmentList | functionCall)
    | columnRef OPERATOR_EQ columnRef (PLUS | MINUS) decimalLiteral
    | columnRef OPERATOR_EQ columnRef (PLUS | MINUS) assignmentSet
    | columnRef OPERATOR_EQ assignmentSet (PLUS | MINUS) columnRef
    | columnRef OPERATOR_EQ columnRef (PLUS | MINUS) assignmentMap
    | columnRef OPERATOR_EQ assignmentMap (PLUS | MINUS) columnRef
    | columnRef OPERATOR_EQ columnRef (PLUS | MINUS) assignmentList
    | columnRef OPERATOR_EQ assignmentList (PLUS | MINUS) columnRef
    | columnRef syntaxBracketLs assignmentIndexKey syntaxBracketRs OPERATOR_EQ constant
    ;

// Index keys for map/list access - supports int, string, boolean, null
assignmentIndexKey
    : decimalLiteral
    | stringLiteral
    | booleanLiteral
    | kwNull
    ;

assignmentSet
    : syntaxBracketLc (assignmentSetElement (syntaxComma assignmentSetElement)*)? syntaxBracketRc
    ;

// Set elements can be constants or nested collections
assignmentSetElement
    : constant
    | assignmentSet
    | assignmentList
    ;

assignmentMap
    : syntaxBracketLc (assignmentMapEntry (syntaxComma assignmentMapEntry)*)? syntaxBracketRc
    ;

// Map entries with flexible key/value types
assignmentMapEntry
    : assignmentMapKey syntaxColon assignmentMapValue
    ;

assignmentMapKey
    : constant
    | assignmentList
    | assignmentSet
    ;

assignmentMapValue
    : constant
    | assignmentSet
    | assignmentList
    | assignmentMap
    ;

assignmentList
    : syntaxBracketLs (assignmentListElement (syntaxComma assignmentListElement)*)? syntaxBracketRs
    ;

// List elements can be constants or nested collections
assignmentListElement
    : constant
    | assignmentSet
    | assignmentList
    | assignmentMap
    ;

assignmentTuple
    : syntaxBracketLr (expression (syntaxComma expression)*) syntaxBracketRr
    ;

insert
    : beginBatch? kwInsert kwInto (keyspace DOT)? table insertColumnSpec? insertValuesSpec ifNotExist? usingTtlTimestamp?
    ;

usingTtlTimestamp
    : kwUsing ttl
    | kwUsing ttl kwAnd timestamp
    | kwUsing timestamp
    | kwUsing timestamp kwAnd ttl
    | kwUsing kwTimeout constant
    | kwUsing ttl kwAnd kwTimeout constant
    | kwUsing timestamp kwAnd kwTimeout constant
    | kwUsing kwTimeout constant kwAnd ttl
    | kwUsing kwTimeout constant kwAnd timestamp
    ;

timestamp
    : kwTimestamp decimalLiteral
    ;

ttl
    : kwTtl decimalLiteral
    ;

usingTimestampSpec
    : kwUsing timestamp
    | kwUsing kwTimeout constant
    | kwUsing timestamp kwAnd kwTimeout constant
    | kwUsing kwTimeout constant kwAnd timestamp
    ;

ifNotExist
    : kwIf kwNot kwExists
    ;

ifExist
    : kwIf kwExists
    ;

insertValuesSpec
    : kwValues '(' expressionList ')'
    | kwJson constant jsonDefault?
    ;

// JSON DEFAULT NULL/UNSET option
jsonDefault
    : kwDefault kwNull
    | kwDefault kwUnset
    ;

insertColumnSpec
    : '(' columnList ')'
    ;

columnList
    : column (syntaxComma column)*
    ;

expressionList
    : expression (syntaxComma expression)*
    ;

expression
    : constant
    | functionCall
    | assignmentMap
    | assignmentSet
    | assignmentList
    | assignmentTuple
    ;

select_
    : kwSelect distinctSpec? kwJson? selectElements fromSpec whereSpec? groupBySpec? orderSpec? perPartitionLimitSpec? limitSpec? allowFilteringSpec? bypassCacheSpec? usingTimeoutSpec?
    ;

allowFilteringSpec
    : kwAllow kwFiltering
    ;

// ScyllaDB: GROUP BY
groupBySpec
    : kwGroup kwBy columnList
    ;

// ScyllaDB: BYPASS CACHE
bypassCacheSpec
    : kwBypass kwCache
    ;

// ScyllaDB: PER PARTITION LIMIT
perPartitionLimitSpec
    : kwPer kwPartition kwLimit decimalLiteral
    ;

// ScyllaDB: USING TIMEOUT
usingTimeoutSpec
    : kwUsing kwTimeout constant
    ;

// ScyllaDB: PRUNE MATERIALIZED VIEW
pruneMaterializedView
    : kwPrune kwMaterialized kwView (keyspace DOT)? materializedView whereSpec? pruneUsingSpec?
    ;

// DESCRIBE statement (server-side since Cassandra 4.0)
describeStatement
    : (kwDescribe | kwDesc) describeTarget describeInternals?
    ;

describeTarget
    : kwCluster
    | kwFull? kwSchema
    | kwKeyspaces
    | kwOnly? kwKeyspace keyspace?
    | kwTables
    | kwTable? (keyspace DOT)? table
    | kwColumnfamily (keyspace DOT)? table
    | kwIndex (keyspace DOT)? OBJECT_NAME
    | kwMaterialized kwView (keyspace DOT)? OBJECT_NAME
    | kwTypes
    | kwType (keyspace DOT)? OBJECT_NAME
    | kwFunctions
    | kwFunction (keyspace DOT)? OBJECT_NAME
    | kwAggregates
    | kwAggregate (keyspace DOT)? OBJECT_NAME
    ;

describeInternals
    : kwWith kwInternals (kwAnd kwPasswords)?
    ;

// Service Level management (ScyllaDB QoS)
serviceLevelName
    : OBJECT_NAME
    | stringLiteral
    ;

serviceLevel
    : kwService kwLevel
    ;

serviceLevels
    : kwService kwLevels
    ;

createServiceLevel
    : kwCreate serviceLevel ifNotExist? serviceLevelName (kwWith propertyList)?
    ;

alterServiceLevel
    : kwAlter serviceLevel serviceLevelName kwWith propertyList
    ;

dropServiceLevel
    : kwDrop serviceLevel ifExist? serviceLevelName
    ;

attachServiceLevel
    : kwAttach serviceLevel serviceLevelName kwTo serviceLevelName
    ;

detachServiceLevel
    : kwDetach serviceLevel kwFrom serviceLevelName
    ;

listServiceLevel
    : kwList serviceLevel serviceLevelName
    | kwList kwAll serviceLevels
    | kwList kwAttached serviceLevel kwOf serviceLevelName
    | kwList kwEffective serviceLevel kwOf serviceLevelName
    ;

propertyList
    : property (kwAnd property)*
    ;

property
    : propertyName OPERATOR_EQ propertyValue
    ;

propertyName
    : OBJECT_NAME
    | kwTimeout
    | kwDefault
    ;

propertyValue
    : constant
    | OBJECT_NAME
    ;

pruneUsingSpec
    : kwUsing kwTimeout constant
    | kwUsing kwConcurrency decimalLiteral
    | kwUsing kwConcurrency decimalLiteral kwAnd kwTimeout constant
    | kwUsing kwTimeout constant kwAnd kwConcurrency decimalLiteral
    ;

limitSpec
    : kwLimit decimalLiteral
    ;

fromSpec
    : kwFrom fromSpecElement
    ;

fromSpecElement
    : OBJECT_NAME
    | reservedKeywordAsTable
    | reservedTypeAsTable
    | OBJECT_NAME '.' OBJECT_NAME
    | OBJECT_NAME '.' reservedKeywordAsTable
    | OBJECT_NAME '.' reservedTypeAsTable
    ;

orderSpec
    : kwOrder kwBy orderSpecElement
    ;

orderSpecElement
    : OBJECT_NAME (kwAsc | kwDesc)?
    | OBJECT_NAME kwAnn kwOf vectorLiteral
    ;

// Vector literal for ANN queries: [0.1, 0.2, 0.3]
vectorLiteral
    : syntaxBracketLs (floatLiteral | decimalLiteral) (syntaxComma (floatLiteral | decimalLiteral))* syntaxBracketRs
    ;

whereSpec
    : kwWhere relationElements
    ;

distinctSpec
    : kwDistinct
    ;

selectElements
    : (star = '*' | selectElement) (syntaxComma selectElement)*
    ;

selectElement
    : OBJECT_NAME '.' '*'
    | columnRef (kwAs OBJECT_NAME)?
    | functionCall (kwAs OBJECT_NAME)?
    | castCall (kwAs OBJECT_NAME)?
    | qualifiedFunctionCall (kwAs OBJECT_NAME)?
    ;

relationElements
    : (relationElement) (kwAnd relationElement)*
    ;

relationElement
    : columnRef (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) constant
    | columnRef (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) functionCall
    | columnRef '.' OBJECT_NAME (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) constant
    | functionCall (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) constant
    | functionCall (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) functionCall
    | columnRef kwIn '(' functionArgs? ')'
    | '(' columnRef (syntaxComma columnRef)* ')' kwIn '(' assignmentTuple (syntaxComma assignmentTuple)* ')'
    | '(' columnRef (syntaxComma columnRef)* ')' (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) (assignmentTuple (syntaxComma assignmentTuple)*)
    | '(' columnRef (syntaxComma columnRef)* ')' (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) scyllaClusteringBound
    | scyllaClusteringBound (OPERATOR_EQ | OPERATOR_LT | OPERATOR_GT | OPERATOR_LTE | OPERATOR_GTE) '(' functionArgs ')'
    | relalationContainsKey
    | relalationContains
    | columnRef kwLike constant
    ;

relalationContains
    : columnRef kwContains constant
    ;

relalationContainsKey
    : columnRef (kwContains kwKey) constant
    ;

// ScyllaDB SCYLLA_CLUSTERING_BOUND for pagination
scyllaClusteringBound
    : K_SCYLLA_CLUSTERING_BOUND '(' functionArgs? ')'
    | K_SCYLLA_CLUSTERING_BOUND
    ;

functionCall
    : OBJECT_NAME '(' STAR ')'
    | OBJECT_NAME '(' functionArgs? ')'
    | K_UUID '(' ')'
    | kwWritetime '(' OBJECT_NAME ')'
    | kwTtl '(' OBJECT_NAME ')'
    | kwToken '(' functionArgs ')'
    ;

// CAST function
castCall
    : kwCast syntaxBracketLr OBJECT_NAME kwAs dataType syntaxBracketRr
    ;

// Qualified function call: keyspace.function(args)
qualifiedFunctionCall
    : OBJECT_NAME '.' OBJECT_NAME '(' functionArgs? ')'
    ;

functionArgs
    : (constant | columnRef | functionCall | qualifiedFunctionCall) (syntaxComma (constant | columnRef | functionCall | qualifiedFunctionCall))*
    ;

constant
    : UUID
    | stringLiteral
    | decimalLiteral
    | floatLiteral
    | hexadecimalLiteral
    | booleanLiteral
    | codeBlock
    | kwNull
    
    | QMARK           // Positional placeholder
    | namedMarker      // Named placeholder :name
    | durationLiteral  // Duration values: 500ms, 1s, etc.
    
    | kwEmpty          // Empty collection literal
    ;

// Named placeholder for prepared statements (:paramName)
namedMarker
    : COLON OBJECT_NAME
    ;

// Duration literal wrapper
durationLiteral
    : DURATION_LITERAL
    ;

decimalLiteral
    : DECIMAL_LITERAL
    ;

floatLiteral
    : DECIMAL_LITERAL
    | FLOAT_LITERAL
    ;

stringLiteral
    : STRING_LITERAL
    ;

booleanLiteral
    : K_TRUE
    | K_FALSE
    ;

hexadecimalLiteral
    : HEXADECIMAL_LITERAL
    ;

keyspace
    : OBJECT_NAME
    | DQUOTE OBJECT_NAME DQUOTE
    ;

table
    : OBJECT_NAME
    | reservedKeywordAsTable
    | reservedTypeAsTable
    ;

column
    : OBJECT_NAME
    | DQUOTE OBJECT_NAME DQUOTE
    | reservedKeywordAsColumn
    ;

// Allow reserved words as unquoted column names (CQL allows this)
reservedKeywordAsColumn
    : K_TIME
    | K_TIMESTAMP
    | K_UUID
    | K_PASSWORD
    | K_TEXT
    | K_KEY
    | K_VALUE
    | K_VALUES
    | K_TYPE
    | K_USER
    | K_ROLE
    | K_STATIC
    
    | K_TWO
    ;

// Column reference in WHERE clauses - allows reserved words as column names
columnRef
    : OBJECT_NAME
    | reservedKeywordAsColumn
    ;

// Allow reserved words as table names (e.g., system_schema.keyspaces)
reservedKeywordAsTable
    : K_KEYSPACES
    | K_TABLES
    | K_COLUMNS
    | K_TYPES
    | K_FUNCTIONS
    | K_AGGREGATES
    | K_VIEWS
    | K_INDEXES
    
    | K_USERS
    ;

// Reserved data type names that can be used as table names
reservedTypeAsTable
    : K_VARCHAR
    | K_TEXT
    | K_INT
    | K_BIGINT
    | K_BOOLEAN
    | K_FLOAT
    | K_DOUBLE
    | K_TWO
    | K_THREE
    | K_ONE
    ;

dataType
    : dataTypeName dataTypeDefinition?
    ;

dataTypeName
    : OBJECT_NAME
    | K_TIMESTAMP
    | K_SET
    | K_ASCII
    | K_BIGINT
    | K_BLOB
    | K_BOOLEAN
    | K_COUNTER
    | K_DATE
    | K_DECIMAL
    | K_DOUBLE
    | K_FLOAT
    | K_FROZEN
    | K_INET
    | K_INT
    | K_LIST
    | K_MAP
    | K_SMALLINT
    | K_TEXT
    | K_TIME
    | K_TIMEUUID
    | K_TINYINT
    | K_TUPLE
    | K_VARCHAR
    | K_VARINT
    | K_TIMESTAMP
    | K_UUID
    
    | K_DURATION       // ScyllaDB
    | K_VECTOR         // ScyllaDB
    ;

dataTypeDefinition
    : syntaxBracketLa dataTypeArg (syntaxComma dataTypeArg)* syntaxBracketRa
    ;

// Data type arguments - supports nested types like frozen<list<int>>, map<int, set<int>>
dataTypeArg
    : dataType
    | decimalLiteral
    ;

orderDirection
    : kwAsc
    | kwDesc
    ;

role
    : OBJECT_NAME
    ;

trigger
    : OBJECT_NAME
    ;

triggerClass
    : stringLiteral
    ;

materializedView
    : OBJECT_NAME
    ;

type_
    : OBJECT_NAME
    ;

aggregate
    : OBJECT_NAME
    ;

function_
    : OBJECT_NAME
    ;

language
    : OBJECT_NAME
    ;

user
    : OBJECT_NAME
    ;

password
    : stringLiteral
    ;

hashKey
    : OBJECT_NAME
    ;

param
    : paramName dataType
    ;

paramName
    : OBJECT_NAME
    | K_INPUT
    ;

kwAdd
    : K_ADD
    ;

kwAggregate
    : K_AGGREGATE
    ;

kwAll
    : K_ALL
    ;

kwAllPermissions
    : K_ALL K_PERMISSIONS
    ;

kwAllow
    : K_ALLOW
    ;

kwAlter
    : K_ALTER
    ;

kwAnd
    : K_AND
    ;

kwApply
    : K_APPLY
    ;

kwAs
    : K_AS
    ;

kwAsc
    : K_ASC
    ;

kwAuthorize
    : K_AUTHORIZE
    ;

kwBatch
    : K_BATCH
    ;

kwBegin
    : K_BEGIN
    ;

kwBy
    : K_BY
    ;

kwCalled
    : K_CALLED
    ;

kwClustering
    : K_CLUSTERING
    ;

kwCompact
    : K_COMPACT
    ;

kwContains
    : K_CONTAINS
    ;

kwCreate
    : K_CREATE
    ;

kwDelete
    : K_DELETE
    ;

kwDesc
    : K_DESC
    ;

kwDescibe
    : K_DESCRIBE
    ;

kwDistinct
    : K_DISTINCT
    ;

kwDrop
    : K_DROP
    ;

kwDurableWrites
    : K_DURABLE_WRITES
    ;

kwEntries
    : K_ENTRIES
    ;

kwExecute
    : K_EXECUTE
    ;

kwExists
    : K_EXISTS
    ;

kwFiltering
    : K_FILTERING
    ;

kwFinalfunc
    : K_FINALFUNC
    ;

kwFrom
    : K_FROM
    ;

kwFull
    : K_FULL
    ;

kwFunction
    : K_FUNCTION
    ;

kwFunctions
    : K_FUNCTIONS
    ;

kwGrant
    : K_GRANT
    ;

kwIf
    : K_IF
    ;

kwIn
    : K_IN
    ;

kwIndex
    : K_INDEX
    ;

kwInitcond
    : K_INITCOND
    ;

kwInput
    : K_INPUT
    ;

kwInsert
    : K_INSERT
    ;

kwInto
    : K_INTO
    ;

kwIs
    : K_IS
    ;

kwJson
    : K_JSON
    ;

kwKey
    : K_KEY
    ;

kwKeys
    : K_KEYS
    ;

kwKeyspace
    : K_KEYSPACE
    ;

kwKeyspaces
    : K_KEYSPACES
    ;

kwLanguage
    : K_LANGUAGE
    ;

kwLimit
    : K_LIMIT
    ;

kwList
    : K_LIST
    ;

kwLogged
    : K_LOGGED
    ;

kwLogin
    : K_LOGIN
    ;

kwMaterialized
    : K_MATERIALIZED
    ;

kwModify
    : K_MODIFY
    ;

kwNosuperuser
    : K_NOSUPERUSER
    ;

kwNorecursive
    : K_NORECURSIVE
    ;

kwNot
    : K_NOT
    ;

kwNull
    : K_NULL
    ;

kwOf
    : K_OF
    ;

kwOn
    : K_ON
    ;

kwOptions
    : K_OPTIONS
    ;

kwOr
    : K_OR
    ;

kwOrder
    : K_ORDER
    ;

kwPassword
    : K_PASSWORD
    ;

kwPrimary
    : K_PRIMARY
    ;

kwRename
    : K_RENAME
    ;

kwReplace
    : K_REPLACE
    ;

kwReplication
    : K_REPLICATION
    ;

kwReturns
    : K_RETURNS
    ;

kwRole
    : K_ROLE
    ;

kwRoles
    : K_ROLES
    ;

kwSelect
    : K_SELECT
    ;

kwSet
    : K_SET
    ;

kwSfunc
    : K_SFUNC
    ;

kwStorage
    : K_STORAGE
    ;

kwStype
    : K_STYPE
    ;

kwSuperuser
    : K_SUPERUSER
    ;

kwTable
    : K_TABLE
    ;

kwTimestamp
    : K_TIMESTAMP
    ;

kwTo
    : K_TO
    ;

kwTrigger
    : K_TRIGGER
    ;

kwTruncate
    : K_TRUNCATE
    ;

kwTtl
    : K_TTL
    ;

kwType
    : K_TYPE
    ;

kwUnlogged
    : K_UNLOGGED
    ;

kwUpdate
    : K_UPDATE
    ;

kwUse
    : K_USE
    ;

kwUser
    : K_USER
    ;

kwUsing
    : K_USING
    ;

kwValues
    : K_VALUES
    ;

kwView
    : K_VIEW
    ;

kwWhere
    : K_WHERE
    ;

kwWith
    : K_WITH
    ;

kwRevoke
    : K_REVOKE
    ;

// BRACKETS
// L - left
// R - right
// a - angle
// c - curly
// r - rounded
syntaxBracketLr
    : LR_BRACKET
    ;

syntaxBracketRr
    : RR_BRACKET
    ;

syntaxBracketLc
    : LC_BRACKET
    ;

syntaxBracketRc
    : RC_BRACKET
    ;

syntaxBracketLa
    : OPERATOR_LT
    ;

syntaxBracketRa
    : OPERATOR_GT
    ;

syntaxBracketLs
    : LS_BRACKET
    ;

syntaxBracketRs
    : RS_BRACKET
    ;

syntaxComma
    : COMMA
    ;

syntaxColon
    : COLON
    ;

// ScyllaDB-specific keyword wrappers
kwBypass    : K_BYPASS;
kwCache     : K_CACHE;
kwTimeout   : K_TIMEOUT;
kwPrune     : K_PRUNE;
kwPer       : K_PER;
kwPartition : K_PARTITION;
kwGroup     : K_GROUP;
kwStatic    : K_STATIC;
kwCast      : K_CAST;
kwLike      : K_LIKE;
kwWritetime : K_WRITETIME;
kwToken     : K_TOKEN;
kwTablets   : K_TABLETS;
kwDefault   : K_DEFAULT;
kwUnset     : K_UNSET;
kwCounter   : K_COUNTER;

// DESCRIBE keywords (only new ones not in base grammar)
kwDescribe    : K_DESCRIBE;
kwCluster     : K_CLUSTER;
kwOnly        : K_ONLY;
kwInternals   : K_INTERNALS;
kwPasswords   : K_PASSWORDS;
kwSchema      : K_SCHEMA;
kwTables      : K_TABLES;
kwTypes       : K_TYPES;
kwAggregates  : K_AGGREGATES;
kwColumnfamily: K_COLUMNFAMILY;

// Service Level keywords
kwService   : K_SERVICE;
kwLevel     : K_LEVEL;
kwLevels    : K_LEVELS;
kwAttach    : K_ATTACH;
kwDetach    : K_DETACH;
kwAttached  : K_ATTACHED;
kwEffective : K_EFFECTIVE;
// Vector Search / ANN keywords
kwCustom    : K_CUSTOM;
kwAnn       : K_ANN;
// Advanced aggregate keywords
kwReducefunc: K_REDUCEFUNC;
// Auth extension keywords
kwNologin   : K_NOLOGIN;
kwUsers     : K_USERS;
kwHashed    : K_HASHED;
// Empty collection literal
kwEmpty     : K_EMPTY;
// PRUNE options
kwConcurrency: K_CONCURRENCY;
// Vector search permission
kwVectorSearchIndexing: K_VECTOR_SEARCH_INDEXING;
