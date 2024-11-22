# Check if Taskfile.yaml exists
if [ ! -f "Taskfile.yaml" ]; then
  echo "Taskfile.yaml not found in the current directory."
  exit 1
fi

# Define content to be appended to 'Taskfile.yaml'
TASKS_YAML_CONTENT=$(cat <<'EOF'
  groovy:author:
    desc: execute Groovy script on AEM author instance
    cmd: |
      SCRIPT="{{.CLI_ARGS}}"
      if [[ ! -f "${SCRIPT}" ]] || [[ "${SCRIPT: -7}" != ".groovy" ]]; then
        echo "Groovy script not found or the file is not a Groovy script: '${SCRIPT}'"
        exit 1
      fi
      RESPONSE=$(curl -u "{{.AEM_AUTHOR_USER}}:{{.AEM_AUTHOR_PASSWORD}}" -k -F "script=@${SCRIPT}" -X POST "{{.AEM_AUTHOR_HTTP_URL}}/bin/groovyconsole/post.json")
      EXCEPTION=$(echo "$RESPONSE" | jq -r '.exceptionStackTrace')
      if [[ $EXCEPTION != "" ]]; then
        echo ""
        echo "Groovy script exception:"
        echo -e "${EXCEPTION}"
        echo ""
      fi
      echo ""
      echo "Groovy script output:"
      OUTPUT=$(echo "${RESPONSE}" | jq -r '.output')
      echo -e "${OUTPUT}"
      echo ""

  groovy:publish:
    desc: execute Groovy script on AEM publish instance
    cmd: |
      SCRIPT="{{.CLI_ARGS}}"
      if [[ ! -f "${SCRIPT}" ]] || [[ "${SCRIPT: -7}" != ".groovy" ]]; then
        echo "Groovy script not found or the file is not a Groovy script: '${SCRIPT}'"
        exit 1
      fi
      RESPONSE=$(curl -u "{{.AEM_PUBLISH_USER}}:{{.AEM_PUBLISH_PASSWORD}}" -k -F "script=@${SCRIPT}" -X POST "{{.AEM_PUBLISH_HTTP_URL}}/bin/groovyconsole/post.json")
      EXCEPTION=$(echo "$RESPONSE" | jq -r '.exceptionStackTrace')
      if [[ $EXCEPTION != "" ]]; then
        echo ""
        echo "Groovy script exception:"
        echo -e "${EXCEPTION}"
        echo ""
      fi
      echo ""
      echo "Groovy script output:"
      OUTPUT=$(echo "${RESPONSE}" | jq -r '.output')
      echo -e "${OUTPUT}"
      echo ""

  crxde:author:
    desc: open CRX/DE on AEM author instance
    cmd: |
      FILE_PATH="{{.CLI_ARGS}}"
      FILE_PATH="${FILE_PATH/.content.xml/jcr:content}"
      FILE_PATH="${FILE_PATH%.xml}"
      REPO_PATH="${FILE_PATH#*jcr_root}"
      REPO_PATH=$(echo "$REPO_PATH" | sed 's/ /%20/g; s/:/%3A/g')
      if [ "{{OS}}" = "windows" ]; then
        start "" "{{.AEM_AUTHOR_HTTP_URL}}/crx/de#${REPO_PATH}"
      elif [ "{{OS}}" = "darwin" ]; then
        open "{{.AEM_AUTHOR_HTTP_URL}}/crx/de#${REPO_PATH}"
      else
        xdg-open "{{.AEM_AUTHOR_HTTP_URL}}/crx/de#${REPO_PATH}"
      fi

  crxde:publish:
    desc: open CRX/DE on AEM publish instance
    cmd: |
      FILE_PATH="{{.CLI_ARGS}}"
      FILE_PATH="${FILE_PATH/.content.xml/jcr:content}"
      FILE_PATH="${FILE_PATH%.xml}"
      REPO_PATH="${FILE_PATH#*jcr_root}"
      REPO_PATH=$(echo "$REPO_PATH" | sed 's/ /%20/g; s/:/%3A/g')
      if [ "{{OS}}" = "windows" ]; then
        start "" "{{.AEM_PUBLISH_HTTP_URL}}/crx/de#${REPO_PATH}"
      elif [ "{{OS}}" = "darwin" ]; then
        open "{{.AEM_PUBLISH_HTTP_URL}}/crx/de#${REPO_PATH}"
      else
        xdg-open "{{.AEM_PUBLISH_HTTP_URL}}/crx/de#${REPO_PATH}"
      fi

EOF
)

# Define IDEA's 'tools/AEM.xml' content
AEM_XML_CONTENT=$(cat <<'EOF'
<toolSet name="AEM">
  <tool name="Content Pull [author]" description="Content Pull [author]" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/aemw" />
      <option name="PARAMETERS" value="content pull -A --path &quot;$FilePath$&quot; --clean" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="Content Pull [publish]" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/aemw" />
      <option name="PARAMETERS" value="content pull -P --path &quot;$FilePath$&quot; --clean" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="Content Clean" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="false" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/aemw" />
      <option name="PARAMETERS" value="content clean --path &quot;$FilePath$&quot;" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="Groovy Console [author]" description="Run currently opened Groovy Script in editor on author instance" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/taskw" />
      <option name="PARAMETERS" value="groovy:author -- $FilePath$" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="Groovy Console [publish]" description="Run currently opened Groovy Script in editor on publish instance" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/taskw" />
      <option name="PARAMETERS" value="groovy:publish -- $FilePath$" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="Content Push [author]" description="Content Push [author]" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/aemw" />
      <option name="PARAMETERS" value="content push -A --path &quot;$FilePath$&quot;" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="Content Push [publish]" description="Content Push [publish]" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/aemw" />
      <option name="PARAMETERS" value="content push -P --path &quot;$FilePath$&quot;" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="CRXDE Open [author]" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="false" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="false">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/taskw" />
      <option name="PARAMETERS" value="crxde:author -- $FilePath$" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
  <tool name="CRXDE Open [publish]" showInMainMenu="false" showInEditor="false" showInProject="false" showInSearchPopup="false" disabled="false" useConsole="false" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="false">
    <exec>
      <option name="COMMAND" value="$ProjectFileDir$/taskw" />
      <option name="PARAMETERS" value="crxde:publish -- $FilePath$" />
      <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
    </exec>
  </tool>
</toolSet>
EOF
)

# Function to create 'tools/AEM.xml' in IntelliJ installations
create_aem_xml() {
  local intellij_dirs=(
    "$HOME/Library/Application Support/JetBrains/IntelliJIdea"*/tools
  )

  for dir in "${intellij_dirs[@]}"; do
    if [ -d "$dir" ]; then
      local aem_file="$dir/AEM.xml"
      echo "Creating AEM.xml in $aem_file"
      echo "$AEM_XML_CONTENT" > "$aem_file"
    fi
  done
}

# Create AEM.xml in IntelliJ installations
create_aem_xml

# Append tasks.yaml content to Taskfile.yaml
echo "Appending tasks.yaml content to Taskfile.yaml"
echo "$TASKS_YAML_CONTENT" >> "Taskfile.yaml"

echo "Script execution completed."
echo "Please restart IntelliJ IDEA to see the changes."
