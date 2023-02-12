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

function dispatchGithubWorkflow() {
  var data = {
    ref: "main",
  };

  const ghToken =
    PropertiesService.getScriptProperties().getProperty("gh_token");

  var options = {
    method: "post",
    contentType: "application/json",
    headers: {
      Accept: "application/vnd.github+json",
      Authorization: `Bearer ${ghToken}`,
      "X-GitHub-Api-Version": "2022-11-28",
      "Content-Type": "application/json",
    },
    payload: JSON.stringify(data),
  };

  UrlFetchApp.fetch(
    "https://api.github.com/repos/james-pickett/gamblebaybirds/actions/workflows/github-pages.yml/dispatches",
    options
  );
}

function mainFunction() {
  const fileCount = countFiles("gamblebaybirds");
  if (fileCount == 0) {
    return;
  }

  console.log("dispatching ");
  dispatchGithubWorkflow();
}
