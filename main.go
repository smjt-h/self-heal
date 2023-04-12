package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/smjt-h/self-heal/model"
)

func main() {
	// Define the file path and the two sets of keywords to search for
	filePath := "/Users/newjoiner1/Desktop/log.txt"
	startKeyword1 := "file name=\""
	endKeyword1 := "\">"
	startKeyword2 := "message="
	endKeyword2 := "/>"

	// Read the file at the specified path
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Convert the file contents to a string
	fileString := string(fileBytes)

	// Find the index of the first set of start and end keywords
	startIndex1 := strings.Index(fileString, startKeyword1)
	endIndex1 := strings.Index(fileString, endKeyword1)

	// Extract the first string between the start and end keywords
	extractedString1 := fileString[startIndex1+len(startKeyword1) : endIndex1]

	// Find the index of the second set of start and end keywords
	startIndex2 := strings.Index(fileString, startKeyword2)
	endIndex2 := strings.Index(fileString, endKeyword2)

	// Extract the second string between the start and end keywords
	extractedString2 := fileString[startIndex2+len(startKeyword2) : endIndex2]

	fmt.Println(extractedString1, extractedString2)
	// Read the file at the first extracted string path
	fileBytes2, err := ioutil.ReadFile(extractedString1)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Print the contents of the second file
	fmt.Println(string(fileBytes2), string(extractedString2))
	heal(extractedString2, string(fileBytes2), extractedString1)
}

func executeBinary(binaryPath string, args ...string) error {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute binary: %v", err)
	}
	return nil
}

func heal(errorMsg, content, filename string) {
	verbose := true

	// content = "/*\n\t* Copyright 2021 Harness Inc. All rights reserved.\n\t* Use of this source code is governed by the PolyForm Free Trial 1.0.0 license\n\t* that can be found in the licenses directory at the root of this repository, also available at\n\t* https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.\n\t*/\n   \n   package io.harness.beans.sweepingoutputs;\n   \n   import static io.harness.annotations.dev.HarnessTeam.CI;\n   \n   import io.harness.annotation.HarnessEntity;\n   import io.harness.annotation.RecasterAlias;\n   import io.harness.annotations.StoreIn;\n   import dev.morphia.utils.Assert;\n   import io.harness.annotations.dev.OwnedBy;\n   import io.harness.beans.build.BuildStatusUpdateParameter;\n   import io.harness.beans.execution.ExecutionSource;\n   import io.harness.beans.steps.CIRegistry;\n   import io.harness.mongo.index.FdIndex;\n   import io.harness.ng.DbAliases;\n   import io.harness.persistence.AccountAccess;\n   import io.harness.persistence.PersistentEntity;\n   import io.harness.persistence.UuidAware;\n   import io.harness.validation.Update;\n   \n   import com.fasterxml.jackson.annotation.JsonIgnoreProperties;\n   import com.fasterxml.jackson.annotation.JsonTypeName;\n   import com.github.reinert.jjschema.SchemaIgnore;\n   import dev.morphia.annotations.Entity;\n   import dev.morphia.annotations.Id;\n   import java.util.List;\n   import javax.validation.constraints.NotNull;\n   import lombok.Builder;\n   import lombok.Data;\n   import org.springframework.data.annotation.TypeAlias;\n   \n   @Data\n   @Builder\n   @JsonIgnoreProperties(ignoreUnknown = true)\n   @StoreIn(DbAliases.CIMANAGER)\n   @Entity(value = \"stageDetails\")\n   @HarnessEntity(exportable = true)\n   @TypeAlias(\"StageDetails\")\n   @JsonTypeName(\"StageDetails\")\n   @OwnedBy(CI)\n   @RecasterAlias(\"io.harness.beans.sweepingoutputs.StageDetails\")\n   public class StageDetails implements PersistentEntity, UuidAware, ContextElement, AccountAccess {\n\t private String stageID;\n\t private String stageRuntimeID;\n\t private BuildStatusUpdateParameter buildStatusUpdateParameter;\n\t private List<CIRegistry> registries;\n\t private long lastUpdatedAt;\n\t private ExecutionSource executionSource;\n\t @Id @NotNull(groups = {Update.class}) @SchemaIgnore private String uuid;\n\t @FdIndex private String accountId;\n   }"
	// errorMsg = "Unused import - dev.morphia.utils.Assert."
	if errorMsg == "" || content == "" {
		log.Fatal("errorMsg or content cannot be empty")
	}
	// Print output

	if verbose {
		fmt.Println("Additional information...")
	}

	code, err := fixCode(content, errorMsg)
	if err != nil {
		log.Fatal(err)
	}

	err = writeToFile(filename, code)
	if err != nil {
		log.Fatal(err)
	}
	// Exit
	os.Exit(0)
}

func fixCode(content, errorMsg string) (string, error) {
	query := "fix the error " + strconv.Quote(errorMsg) + "\n\tand give me the updated code from" + strconv.Quote(content) + "\n### Fixed code"

	fmt.Println("-----------\n\n\n")
	fmt.Println(query, os.Getenv("OPENAI_TOKEN"))

	code, err := openAIRequest(query)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return code.Choices[0].Text, nil
}

func openAIRequest(query string) (*model.CompletionResponse, error) {
	token := os.Getenv("OPENAI_TOKEN")

	url := "https://api.openai.com/v1/completions"
	method := "POST"

	// fmt.Println("query==", query)
	payload := strings.NewReader(`{
	"model": "text-davinci-003",
	"prompt": ` + strconv.Quote(query) + `,
	"temperature": 0,
	"max_tokens": 2000,
	"top_p": 1.0,
	"stop": ["###"],
	"frequency_penalty": 0.0,
	"presence_penalty": 0.0
  }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var completionResp model.CompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return &completionResp, nil
}

func writeToFile(filename, code string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s\n", code))
	if err != nil {
		return err
	}

	fmt.Printf("Code successfully written to file '%s'.\n", filename)
	return nil
}
