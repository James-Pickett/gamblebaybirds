function countFiles(folderName) {
  const folderId = DriveApp.getFoldersByName(folderName).next().getId();
  const folder = DriveApp.getFolderById(folderId);
  const files = folder.getFiles();
  let count = 0;
  while (files.hasNext()) {
    files.next();
    count++;
  }
  return count;
}

function checkProperty(propertyName, newProperty) {
  newProperty = newProperty.toString();
  const oldProperty =
    PropertiesService.getScriptProperties().getProperty(propertyName);
  PropertiesService.getScriptProperties().setProperty(
    propertyName,
    newProperty
  );

  if (newProperty == "0") {
    return false;
  }

  return newProperty != oldProperty;
}

function mainFunction() {
  const fileCount = countFiles("gamblebaybirds");
  const runCode = checkProperty("file_count", fileCount);

  if (runCode) {
    // here execute your main code
    //
    console.log("I am executed!");
    //
  }
}
