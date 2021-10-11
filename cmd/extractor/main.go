package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"os"
)

// Column represents a bounding box
type Column struct {
	Left float64
	Right float64
}

// Line is text within a column
type Line struct {
	Index int
	Text string
}

// extract takes the output from Textract and returns an ordered string
func extract(response *textract.GetDocumentTextDetectionOutput) (output []string) {
	columns := []*Column{}
	lines := []*Line{}

	// Adapted from https://github.com/aws-samples/amazon-textract-code-samples/blob/master/python/03-reading-order.py
	// We have to identify separate columns, then identify whether a given line's bounding box falls within the column
	for _, item := range response.Blocks {
		if *item.BlockType == "LINE" {
			var columnFound bool
			for index, column := range columns {
				bboxLeft := *item.Geometry.BoundingBox.Left
				bboxRight := *item.Geometry.BoundingBox.Left + *item.Geometry.BoundingBox.Width
				bboxCenter := *item.Geometry.BoundingBox.Left + *item.Geometry.BoundingBox.Width/2
				columnCenter := column.Left + column.Right/2

				if (bboxCenter > column.Left && bboxCenter < column.Right) || (columnCenter > bboxLeft && columnCenter < bboxRight) {
					// Bounding box appears inside the column
					lines = append(lines, &Line{Index: index, Text: *item.Text})
					columnFound = true
					break
				}
			}
			if columnFound == false {
				column := &Column{
					Left:  *item.Geometry.BoundingBox.Left,
					Right: (*item.Geometry.BoundingBox.Left + *item.Geometry.BoundingBox.Width),
				}
				columns = append(columns, column)
				lines = append(lines, &Line{Index: len(columns) - 1, Text: *item.Text})
			}
		}
	}

	// Append lines in order
	for index, _ := range columns {
		for _, line := range lines {
			if index == line.Index {
				output = append(output, line.Text)
			}
		}
	}
	return output
}

func main() {
	// AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY env variables required
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	svc := textract.New(sess)

	// Hold the entire output of the PDF document
	documentText := []string{}

	// Job IDs
	// jobId := "4b3edd39e7a604518b521c7ee4456b9c0b8d9a2e974368d25ebb6d8bbda82b21" # Page 5 only
	// jobId := "3e12ae2e58079968e7e3d703b01d0af8ebf0a4a0089a6dcac104dc6198343a47" # Whole PDF
	jobId := "3e12ae2e58079968e7e3d703b01d0af8ebf0a4a0089a6dcac104dc6198343a47"
	maxResults := int64(1000)

	job := &textract.GetDocumentTextDetectionInput{
		JobId:      &jobId,
		MaxResults: &maxResults,
		NextToken:  nil,
	}

	// Get the initial response and see if there is further output to read
	response, err := svc.GetDocumentTextDetection(job)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	output := extract(response)
	documentText = append(documentText, output...)

	nextToken := response.NextToken
	for {
		if nextToken == nil {
			break
		}
		job := &textract.GetDocumentTextDetectionInput{
			JobId:      &jobId,
			MaxResults: &maxResults,
			NextToken:  nextToken,
		}
		response, err := svc.GetDocumentTextDetection(job)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		output := extract(response)
		documentText = append(documentText, output...)
		nextToken = response.NextToken
	}

	// Write the output to a file
	f, err := os.Create("rrg.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	for _, line := range documentText {
		_, err = f.WriteString(fmt.Sprintf("%s\n", line))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
