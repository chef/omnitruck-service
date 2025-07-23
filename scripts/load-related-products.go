package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/xuri/excelize/v2"
)

// Reads AWS credentials from the default credentials file (~/.aws/credentials) using AWS SDK v2
func getAWSDynamoDBClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// Uploads the BOM to products mapping to DynamoDB
func uploadToDynamoDB(tableName string, bomToProducts map[string]map[string]string) error {
	client, err := getAWSDynamoDBClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for bom, products := range bomToProducts {
		// Convert products map into DynamoDB map
		ddbProducts := make(map[string]types.AttributeValue)
		for k, v := range products {
			ddbProducts[k] = &types.AttributeValueMemberS{Value: v}
		}
		input := &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]types.AttributeValue{
				"bom":      &types.AttributeValueMemberS{Value: bom},
				"products": &types.AttributeValueMemberM{Value: ddbProducts},
			},
		}
		_, err = client.PutItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to put item for BOM %s: %v", bom, err)
		}
	}
	return nil
}

// Reads the Excel sheet and returns a map of BOM to products with correct UI name
func readBomProductMap(f *excelize.File, sheetName string) (map[string]map[string]string, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	if len(rows) < 4 {
		return nil, fmt.Errorf("not enough rows in sheet")
	}

	// Use 3rd row (index 2) as header
	header := rows[2]
	var (
		softwareCol, uiNewCol, uiOldCol, mappedSkusCol int = -1, -1, -1, -1
	)
	for i, col := range header {
		switch col {
		case "Software":
			softwareCol = i
		case "UI - Presentable Software Name (NEW)":
			uiNewCol = i
		case "UI - Presentable Software Name (OLD)":
			uiOldCol = i
		case "Mapped SKUs":
			mappedSkusCol = i
		}
	}
	if softwareCol == -1 {
		return nil, fmt.Errorf("software column not found")
	}
	if mappedSkusCol == -1 {
		return nil, fmt.Errorf("Mapped SKUs column not found")
	}

	// All columns after 'Mapped SKUs' are BOMs
	result := make(map[string]map[string]string)
	for col := mappedSkusCol + 1; col < len(header); col++ {
		bom := strings.TrimSpace(header[col])
		if bom == "" {
			continue
		}
		products := make(map[string]string)
		for rowIdx := 3; rowIdx < len(rows); rowIdx++ {
			row := rows[rowIdx]
			if col < len(row) && (row[col] == "x" || row[col] == "X") {
				if softwareCol < len(row) {
					// Truncate at the first space, then remove all remaining spaces for softwareKey
					rawSoftware := row[softwareCol]
					// Truncate at first space (remove everything after first space)
					if idx := strings.Index(rawSoftware, " "); idx != -1 {
						rawSoftware = rawSoftware[:idx]
					}
					softwareKey := rawSoftware
					var uiName string
					if uiNewCol != -1 && uiNewCol < len(row) && row[uiNewCol] != "" {
						uiName = row[uiNewCol]
					} else if uiOldCol != -1 && uiOldCol < len(row) && row[uiOldCol] != "" {
						uiName = row[uiOldCol]
					} else {
						uiName = softwareKey
					}
					products[softwareKey] = uiName
				}
			}
		}
		// Always add the BOM, even if products is empty
		result[bom] = products
	}
	return result, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the Excel file name or full path: ")
	excelFile, _ := reader.ReadString('\n')
	excelFile = strings.TrimSpace(excelFile)
	if excelFile == "" {
		log.Fatalf("Excel file name cannot be empty")
	}

	// Check if the file exists
	if _, err := os.Stat(excelFile); os.IsNotExist(err) {
		log.Fatalf("The file %s was not found. Please provide the correct path.", excelFile)
	}
	sheetName := "SKU Data - Software " // Change if your sheet name is different

	// Open Excel file
	f, err := excelize.OpenFile(excelFile)
	if err != nil {
		log.Fatalf("Failed to open Excel file: %v", err)
	}

	bomToProducts, err := readBomProductMap(f, sheetName)
	if err != nil {
		log.Fatalf("Failed to map BOM to products: %v", err)
	}

	// Prompt user for DynamoDB table name
	fmt.Print("Enter DynamoDB table name to upload data: ")
	tableName, _ := reader.ReadString('\n')
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		log.Fatalf("Table name cannot be empty")
	}

	// Upload to DynamoDB
	err = uploadToDynamoDB(tableName, bomToProducts)
	if err != nil {
		log.Fatalf("Failed to upload to DynamoDB: %v", err)
	}
	fmt.Println("Upload to DynamoDB successful.")

}
