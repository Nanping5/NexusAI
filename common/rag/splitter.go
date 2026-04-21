package rag

import (
	"NexusAi/pkg/utils"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type ContentType int

const (
	ContentTypeParagraph ContentType = iota
	ContentTypeHeading
	ContentTypeCode
	ContentTypeList
	ContentTypeTable
	ContentTypeQuote
	ContentTypeEmpty
)

func (c ContentType) String() string {
	switch c {
	case ContentTypeParagraph:
		return "paragraph"
	case ContentTypeHeading:
		return "heading"
	case ContentTypeCode:
		return "code"
	case ContentTypeList:
		return "list"
	case ContentTypeTable:
		return "table"
	case ContentTypeQuote:
		return "quote"
	case ContentTypeEmpty:
		return "empty"
	default:
		return "unknown"
	}
}

type semanticBlock struct {
	Content   string
	Type      ContentType
	StartLine int
	Heading   string
}

type semanticChunk struct {
	Content     string
	ChunkIndex  int
	StartLine   int
	EndLine     int
	ContentType ContentType
}

func detectContentType(line string) ContentType {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return ContentTypeEmpty
	}

	if strings.HasPrefix(trimmed, "#") {
		rest := strings.TrimLeft(trimmed, "#")
		if len(rest) > 0 && rest[0] == ' ' {
			return ContentTypeHeading
		}
	}

	if strings.HasPrefix(trimmed, "```") {
		return ContentTypeCode
	}

	if strings.HasPrefix(trimmed, "> ") {
		return ContentTypeQuote
	}

	if matched, _ := regexp.MatchString(`^[\-\*]\s`, trimmed); matched {
		return ContentTypeList
	}
	if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
		return ContentTypeList
	}

	if strings.Contains(trimmed, "|") && strings.Count(trimmed, "|") >= 2 {
		return ContentTypeTable
	}

	return ContentTypeParagraph
}

func estimateTokens(text string) int {
	return len(text) / 3
}

func splitIntoSemanticBlocks(content string) []semanticBlock {
	lines := strings.Split(content, "\n")
	var blocks []semanticBlock
	var currentBlock strings.Builder
	currentType := ContentTypeParagraph
	startLine := 1
	currentHeading := ""

	flushBlock := func() {
		if currentBlock.Len() > 0 {
			blocks = append(blocks, semanticBlock{
				Content:   currentBlock.String(),
				Type:      currentType,
				StartLine: startLine,
				Heading:   currentHeading,
			})
			currentBlock.Reset()
		}
	}

	for i, line := range lines {
		lineNum := i + 1
		typ := detectContentType(line)

		switch typ {
		case ContentTypeEmpty:
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
			}
			continue

		case ContentTypeHeading:
			flushBlock()
			currentBlock.WriteString(line)
			currentType = ContentTypeHeading
			startLine = lineNum
			currentHeading = strings.TrimPrefix(strings.TrimSpace(line), "# ")
			flushBlock()
			continue

		case ContentTypeCode:
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") && currentType == ContentTypeCode {
				currentBlock.WriteString("\n")
				currentBlock.WriteString(line)
				flushBlock()
				continue
			}
			if currentType != ContentTypeCode {
				flushBlock()
				currentType = ContentTypeCode
				startLine = lineNum
			}
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
			}
			currentBlock.WriteString(line)
			continue

		case ContentTypeList, ContentTypeTable, ContentTypeQuote:
			if currentType != typ {
				flushBlock()
				currentType = typ
				startLine = lineNum
			}
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
			}
			currentBlock.WriteString(line)
			continue

		default:
			if currentType == ContentTypeCode {
				if currentBlock.Len() > 0 {
					currentBlock.WriteString("\n")
				}
				currentBlock.WriteString(line)
				continue
			}
			if currentType != ContentTypeParagraph {
				flushBlock()
				currentType = ContentTypeParagraph
				startLine = lineNum
			}
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
			}
			currentBlock.WriteString(line)
		}
	}

	flushBlock()
	return blocks
}

func mergeIntoSlidingWindow(
	blocks []semanticBlock,
	maxTokens int,
	overlapTokens int,
) []semanticChunk {
	if len(blocks) == 0 {
		return nil
	}

	var chunks []semanticChunk
	var currentContent strings.Builder
	currentType := ContentTypeParagraph
	chunkIndex := 0
	startLine := 1

	overlapContent := ""

	flushCurrentChunk := func() {
		if currentContent.Len() == 0 {
			return
		}
		endLine := startLine + strings.Count(currentContent.String(), "\n")
		chunks = append(chunks, semanticChunk{
			Content:     currentContent.String(),
			ChunkIndex:  chunkIndex,
			StartLine:   startLine,
			EndLine:     endLine,
			ContentType: currentType,
		})
		overlapContent = currentContent.String()
		if len(overlapContent) > overlapTokens*3 {
			overlapContent = overlapContent[len(overlapContent)-overlapTokens*3:]
		}
		currentContent.Reset()
		chunkIndex++
	}

	for _, block := range blocks {
		blockTokens := estimateTokens(block.Content)

		if blockTokens > maxTokens {
			flushCurrentChunk()

			subChunks := splitOversizedBlock(block, maxTokens, overlapTokens)
			for j, sc := range subChunks {
				sc.ChunkIndex = chunkIndex + j
				chunks = append(chunks, sc)
			}
			chunkIndex += len(subChunks)
			if len(subChunks) > 0 {
				overlapContent = subChunks[len(subChunks)-1].Content
				if len(overlapContent) > overlapTokens*3 {
					overlapContent = overlapContent[len(overlapContent)-overlapTokens*3:]
				}
			}
			continue
		}

		candidateTokens := estimateTokens(currentContent.String()) + blockTokens
		if currentContent.Len() > 0 && candidateTokens > maxTokens {
			flushCurrentChunk()

			if overlapContent != "" {
				overlapStr := overlapContent
				if len(overlapStr) > overlapTokens*3 {
					overlapStr = overlapStr[len(overlapStr)-overlapTokens*3:]
				}
				currentContent.WriteString(overlapStr)
				if len(chunks) > 0 {
					startLine = chunks[len(chunks)-1].StartLine
				}
			}
		}

		if currentContent.Len() > 0 {
			currentContent.WriteString("\n\n")
		}
		currentContent.WriteString(block.Content)
		currentType = block.Type
		if startLine == 0 {
			startLine = block.StartLine
		}
		if startLine > block.StartLine {
			startLine = block.StartLine
		}
	}

	flushCurrentChunk()
	return chunks
}

func splitOversizedBlock(block semanticBlock, maxTokens int, overlapTokens int) []semanticChunk {
	lines := strings.Split(block.Content, "\n")
	var chunks []semanticChunk
	var current strings.Builder
	startLine := block.StartLine
	chunkIndex := 0

	avgTokensPerLine := 20
	if len(lines) > 0 {
		avgTokensPerLine = estimateTokens(block.Content) / len(lines)
		if avgTokensPerLine == 0 {
			avgTokensPerLine = 20
		}
	}

	linesPerChunk := maxTokens / avgTokensPerLine
	if linesPerChunk < 3 {
		linesPerChunk = 3
	}

	for i, line := range lines {
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)

		lineCount := strings.Count(current.String(), "\n") + 1
		currentTokens := estimateTokens(current.String())

		if lineCount >= linesPerChunk || currentTokens >= maxTokens || i == len(lines)-1 {
			chunks = append(chunks, semanticChunk{
				Content:     current.String(),
				ChunkIndex:  chunkIndex,
				StartLine:   startLine,
				EndLine:     startLine + lineCount - 1,
				ContentType: block.Type,
			})
			startLine += lineCount
			chunkIndex++
			current.Reset()
		}
	}

	return chunks
}

func SplitDocument(content string, filePath string, chunkSize int, overlapSize int) []*schema.Document {
	if chunkSize <= 0 {
		chunkSize = 800
	}
	if overlapSize < 0 {
		overlapSize = 0
	}

	blocks := splitIntoSemanticBlocks(content)

	totalTokens := estimateTokens(content)
	if totalTokens <= chunkSize {
		id, _ := utils.GenerateShortID(10)
		return []*schema.Document{{
			ID:       id,
			Content:  content,
			MetaData: map[string]any{
				"source": filePath,
				"chunk":  0,
			},
		}}
	}

	semanticChunks := mergeIntoSlidingWindow(blocks, chunkSize, overlapSize)

	var docs []*schema.Document
	for _, sc := range semanticChunks {
		id, _ := utils.GenerateShortID(10)
		docs = append(docs, &schema.Document{
			ID:       id,
			Content:  sc.Content,
			MetaData: map[string]any{
				"source":       filePath,
				"chunk":        sc.ChunkIndex,
				"content_type": sc.ContentType.String(),
			},
		})
	}

	return docs
}
