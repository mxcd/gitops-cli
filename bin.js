#!/usr/bin/env node

// Imports
import {ArgumentParser} from "argparse";
import {patch as patchGitOpsCli} from './index.js'


// version (same as package.json due to npx problems)
const version = "1.0.0";
 
const parser = new ArgumentParser({
  description: 'gitops-cli'
});

/* 
variables that can be given as environment variables
- GITOPS_BASE_URL: base URL to the git system (e.g. https://gitlab.com, https://github.com)
- GITOPS_AT: access token to the git system
- GITOPS_REPO: gitops repo to use (e.g. '/foobar/gitops')
- GITOPS_BRANCH: gitops branch to use (defaults to 'master')
- GITOPS_APPLICATIONS_DIR: directory where the applications are stored (defaults to 'applications')
- GITOPS_VALUES_FILE: file containing the values to use (defaults to 'values.yaml')
*/

const ENVIRONMENT_VARIABLES = [
  {name: 'GITOPS_BASE_URL', default: 'https://github.com', varialbe: 'url', required: true},
  {name: 'GITOPS_AT', default: null, varialbe: 'access_token', required: true},
  {name: 'GITOPS_REPO', default: null, varialbe: 'repo', required: true},
  {name: 'GITOPS_BRANCH', varialbe: 'branch'},
  {name: 'GITOPS_APPLICATIONS_DIR', default: 'applications', varialbe: 'applications_dir', required: true},
  {name: 'GITOPS_VALUES_FILE', default: 'values.yaml', varialbe: 'values_file', required: true},
]

const ACTIONS = [
  {name: 'patch', parameters: ['application', 'patch_field', 'patch_value'], callback: patchGitOpsCli},
]

parser.add_argument('action', {metavar: 'action', type: String, nargs: '?', help: 'action to be executed'});
parser.add_argument('parameters', {metavar: 'parameters', type: String, nargs: '*', help: 'action parameters to be used'});
parser.add_argument('-v', '--version', { action: 'version', version });
parser.add_argument('-t', '--access-token', {type: String, help: 'access token with API permissions'});
parser.add_argument('-u', '--url', {type: String, default: 'https://github.com', help: 'git system base url (e.g. https://github.com)'})
parser.add_argument('--verbose', {action: 'store_true', help: 'increased console output'})
parser.add_argument('--branch', {metavar: 'branch', type: String, default: 'master', help: 'gitops branch to use'})
parser.add_argument('--repo', {metavar: 'repo', type: String, help: 'gitops repo to use'})
parser.add_argument('--applications-dir', {type: String, default: 'applications', help: 'applications directory in gitops repo'})
parser.add_argument('--values-file', {type: String, default: 'values.yaml', help: 'values file to patch'})

const args = parser.parse_args()

for(const variable of ENVIRONMENT_VARIABLES) {
  if(process.env[variable.name] && !args[variable.varialbe]) {
    args[variable.varialbe] = process.env[variable.name]
  }

  if(variable.required && !args[variable.varialbe]) {
    console.error(`The variable '${variable.name}' is required. Either set it as an environment variable or pass it as an argument.`)
    process.exit(1)
  }
}

if(args.verbose) {
  console.dir(args);
}

if(!args.action) {
  console.error('ERROR: no action given')
  process.exit(1)
}

(async () => {
  for(const action of ACTIONS) {
    if(args.action === action.name && args.parameters.length === action.parameters.length) {
      if(args.verbose) console.log("performing action 'patch'");

      let options = {
        ...args
      }

      for(let i = 0; i < action.parameters.length; ++i) {
        options[action.parameters[i]] = args.parameters[i]
      }

      try {
        await action.callback(options)
        process.exit(0)
      }
      catch(error) {
        console.error(`ERROR: ${error.message}`)
        process.exit(1)
      }
    }
  }
  
  // only reached if no action was found
  console.error(`ERROR: action '${args.action}' with ${args.parameters.length} parameters not allowed`)
})();
