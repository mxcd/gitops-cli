import {
  apiTest as gitlabApiTest,
  getApiDriver as getGitlabApiDriver,
} from "gitlab-x";
import {
  apiTest as githubApiTest,
  getApiDriver as getGithubApiDriver,
} from "github-x";
import { load as loadYml, dump as dumpYml } from "js-yaml";

/* 
gitops patch function
looks up a file and creates a new commit, patching the defined field in the process
required options:
  - url: base url of the git system
  - access_token: access token with API permissions
  - repo: git repository to use
  - branch: branch to use in git repository. defaults to repo's default branch if not given
  - behalf: user to act on behalf of
  - applications_dir: directory where the applications are stored
  - values_file: file containing the value to patch
  - application: name of the application to patch
  - patch_field: field to patch
  - patch_value: value to patch
  
*/
export async function patch(options) {
  const verbose = options.verbose;

  if (verbose)
    console.log(`patch options: ${JSON.stringify(options, null, 2)}`);

  if (
    !(
      options.values_file.endsWith(".yml") ||
      options.values_file.endsWith(".yaml")
    )
  ) {
    throw new Error(`values_file must be a *.yml or *.yaml file`);
  }

  let gitProvider = await getGitProvider(options);
  if (verbose) console.log(`Git Provider: ${gitProvider}`);
  let apiDriver;
  if (gitProvider === "gitlab") {
    apiDriver = getGitlabApiDriver(options);
  } else if (gitProvider === "github") {
    apiDriver = getGithubApiDriver(options);
  } else {
    throw new Error(`Unsupported git provider: ${gitProvider}`);
  }

  if (!(await apiDriver.getVersion())) {
    throw new Error(`API check failed`);
  }

  const projectIdentifier = `/${stripSlashes(options.repo)}`;

  let filePath = `/${stripSlashes(options.applications_dir)}/${stripSlashes(
    options.application
  )}/${stripSlashes(options.values_file)}`;
  if (verbose) console.log(`file path: ${filePath}`);
  const fileExists = await apiDriver.fileExists(
    projectIdentifier,
    filePath,
    options.branch
  );

  if (!fileExists) {
    if (verbose)
      console.log(
        `file '${filePath}' does not exist in branch '${options.branch}'`
      );
    filePath = `/${stripSlashes(options.applications_dir)}/${stripSlashes(
      options.application
    )}/${stripSlashes(toggleYamlFileExtension(options.values_file))}`;
    if (verbose)
      console.log(`trying opposite yaml file extension '${filePath}'`);
    const toggledYamlFileExists = await apiDriver.fileExists(
      projectIdentifier,
      filePath,
      options.branch
    );
    if (!toggledYamlFileExists) {
      throw new Error(`file '${filePath}' does not exist`);
    }
  }

  if (verbose) console.log(`file '${filePath}' exists`);

  const fileContent = await apiDriver.getRawFile(
    projectIdentifier,
    filePath,
    options.branch
  );
  if (verbose) console.log(`file contents: \n\n----\n${fileContent}\n----\n\n`);

  const yml = loadYml(fileContent);
  let patchFields = [];
  if (options.patch_field.startsWith(".")) {
    if (verbose) console.log(`patch field is assumed to be a yaml path`);
    patchFields = [options.patch_field];
  } else {
    if (verbose) console.log(`patch field is assumed to be a field name`);
    patchFields = findYmlField(yml, options.patch_field);
    if (verbose) console.log(`found patch fields: ${patchFields}`);
  }

  let changes = false;
  for (const patchField of patchFields) {
    if (!existsYmlField(yml, patchField)) {
      throw new Error(
        `field '${patchField}' does not exist in file '${filePath}'`
      );
    }

    if (verbose) console.log(`patching field '${patchField}'`);
    const patchValue = getYmlFieldValue(yml, patchField);
    if (verbose) console.log(`old value: ${patchValue}`);
    const newValue = options.patch_value;
    if (verbose) console.log(`new value: ${newValue}`);
    if (patchValue !== newValue) {
      changes = true;
      setYmlFieldValue(yml, patchField, newValue);
    }
  }

  if (!changes) {
    console.log(`no changes to commit`);
    return;
  }

  const ymlDump = dumpYml(yml);

  if (verbose) console.log(`new file contents: \n\n----\n${ymlDump}\n----\n\n`);

  if (verbose) console.log("creating commit");
  let targetBranch = options.branch;
  if (typeof ref === "undefined") {
    // TODO add to defaultBranch function to gitlab-x
    const defaultBranch = (await apiDriver.getProject(projectIdentifier))
      .default_branch;
    if (verbose)
      console.log(
        `'ref' is not specified. Using default branch '${defaultBranch}'`
      );
    targetBranch = defaultBranch;
  }

  let commitMessage = `Patched '${filePath}'`;
  if(options.behalf) {
    commitMessage = `[${options.behalf}] ${commitMessage}`;
  }
  if(options.message) {
    commitMessage = options.message;
  }

  let commitObject = {
    branch: targetBranch,
    commit_message: commitMessage,
    actions: [
      {
        action: "update",
        file_path: filePath,
        content: ymlDump,
        encoding: "text",
      },
    ],
  };
  await apiDriver.postCommit(projectIdentifier, commitObject);
  if (verbose) console.log("commit done");
  return;
}

/* 
checks if the options indicate a github or gitlab provider
required options:
  - url
  - access_token
returns
  - 'github' if the base url points to a github url with activated api
  - 'gitlab' if the base url points to a gitlab url with activated api
*/
async function getGitProvider(options) {
  if (options.verbose) console.log("checking for gitlab api");
  try {
    const gitlabApiResult = await gitlabApiTest(options);
    if (gitlabApiResult) {
      if (options.verbose) console.log("found gitlab api");
      return "gitlab";
    }
  } catch (e) {
    // error is expected if the gitlab api is not present
  }

  try {
    if (options.verbose) console.log("checking for github api");
    const githubApiResult = await githubApiTest(options);
    if (githubApiResult) {
      if (options.verbose) console.log("found github api");
      return "github";
    }
  } catch (e) {
    // error is expected if the github api is not present
  }
  return "";
}

// strips tailing and leading slashes from a string
function stripSlashes(str) {
  return str.replace(/^\/|\/$/g, "");
}

function toggleYamlFileExtension(fileName) {
  if (fileName.endsWith(".yml")) {
    return fileName.replace(".yml", ".yaml");
  } else if (fileName.endsWith(".yaml")) {
    return fileName.replace(".yaml", ".yml");
  } else {
    return fileName;
  }
}

function findYmlField(yml, fieldName, subPath = "") {
  let occurences = [];
  for (let key in yml) {
    if (key === fieldName) {
      occurences.push(`${subPath}.${key}`);
    } else if (typeof yml[key] === "object") {
      occurences = occurences.concat(
        findYmlField(yml[key], fieldName, `${subPath}.${key}`)
      );
    }
  }
  return occurences;
}

function existsYmlField(yml, fieldPath) {
  const fieldPathParts = fieldPath.split(".");
  let currentYml = yml;
  for (let i = 0; i < fieldPathParts.length; i++) {
    const fieldPathPart = fieldPathParts[i];
    if (fieldPathPart === "") continue;
    if (!currentYml[fieldPathPart]) {
      return false;
    }
    currentYml = currentYml[fieldPathPart];
  }
  return true;
}

function getYmlFieldValue(yml, fieldPath) {
  const fieldPathParts = fieldPath.split(".");
  let currentYml = yml;
  for (let i = 0; i < fieldPathParts.length; i++) {
    const fieldPathPart = fieldPathParts[i];
    if (fieldPathPart === "") continue;
    currentYml = currentYml[fieldPathPart];
  }
  return currentYml;
}

function setYmlFieldValue(yml, fieldPath, value) {
  const fieldPathParts = fieldPath.split(".");
  let currentYml = yml;
  for (let i = 0; i < fieldPathParts.length; i++) {
    const fieldPathPart = fieldPathParts[i];
    if (fieldPathPart === "") continue;
    if (i === fieldPathParts.length - 1) {
      currentYml[fieldPathPart] = value;
    } else {
      if (!currentYml[fieldPathPart]) {
        currentYml[fieldPathPart] = {};
      }
      currentYml = currentYml[fieldPathPart];
    }
  }
  return yml;
}
