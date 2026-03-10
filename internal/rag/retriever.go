package rag

import (
	"context"

	"milvus-kb-demo/internal/llm"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const Dim = 2048

type Retriever struct {
	Milvus         client.Client
	CollectionName string
	LLM            *llm.Client
}

func NewRetriever(milvus client.Client, collectionName string, llmClient *llm.Client) *Retriever {
	return &Retriever{
		Milvus:         milvus,
		CollectionName: collectionName,
		LLM:            llmClient,
	}
}

func (r *Retriever) Search(ctx context.Context, question string, topK int) ([]string, error) {
	if topK < 3 {
		topK = 3
	}
	if topK > 5 {
		topK = 5
	}
	qVec, err := r.LLM.Embed(ctx, question)
	if err != nil {
		return nil, err
	}
	sp, _ := entity.NewIndexFlatSearchParam()
	res, err := r.Milvus.Search(ctx, r.CollectionName, []string{}, "", []string{"content"},
		[]entity.Vector{entity.FloatVector(qVec)}, "vector", entity.L2, topK, sp)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	col := res[0].Fields.GetColumn("content")
	contentCol, ok := col.(*entity.ColumnVarChar)
	if !ok {
		return nil, nil
	}
	docs := make([]string, 0, res[0].ResultCount)
	for i := 0; i < res[0].ResultCount; i++ {
		docs = append(docs, contentCol.Data()[i])
	}
	return docs, nil
}
