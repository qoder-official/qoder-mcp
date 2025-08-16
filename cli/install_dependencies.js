const fs = require('fs-extra');
const path = require('path');
const shell = require('shelljs');

// Derive root paths relative to this file so that the module is relocatable
const projectRoot = path.resolve(__dirname, '..');
const serversRoot = path.join(projectRoot, 'servers');

/**
 * Install / tidy dependencies for a single MCP server if required.
 * Mirrors the behaviour that previously lived in cli.js but without relying on
 * outer-scope variables, making the function easier to test and reuse.
 *
 * @param {{id: string, config: any}} server – server descriptor returned from getServers()
 * @param {(msg: string) => void} log – logger function (e.g. console.log or wrapper)
 */
async function installDependencies(server, log) {
  const { id, config } = server;
  const serverDir = path.join(serversRoot, id);
  const executionCwd = config.working_dir
    ? path.join(serverDir, config.working_dir)
    : serverDir;
  const lang = config.lang;

  log(`\n📦 Installing dependencies for '${id}' (${lang})...`);

  try {
    if (lang === 'node') {
      const pkgPath = path.join(executionCwd, 'package.json');
      if (await fs.pathExists(pkgPath)) {
        const pkgJson = await fs.readJson(pkgPath);
        if (pkgJson.dependencies || pkgJson.devDependencies) {
          let installFlags = '--omit=dev';
          // If TypeScript is a dev dependency, we need it during build steps.
          if (pkgJson.devDependencies?.typescript) {
            installFlags = '';
          }
          log(`   Running npm install ${installFlags}...`);
          const npmCmd = process.platform === 'win32' ? 'npm.cmd' : 'npm';
          const installRes = shell.exec(`${npmCmd} install ${installFlags}`, {
            cwd: executionCwd,
            silent: true,
          });
          if (installRes.code !== 0) {
            log(`   ❌ npm install failed for '${id}'.`);
            log(installRes.stderr);
          } else {
            log('   ✅ npm dependencies installed.');
          }

          if (pkgJson.scripts?.build) {
            log('   Running npm run build...');
            const buildRes = shell.exec(`${npmCmd} run build`, {
              cwd: executionCwd,
              silent: true,
            });
            if (buildRes.code !== 0) {
              log(`   ❌ npm run build failed for '${id}'.`);
              log(buildRes.stderr);
            } else {
              log('   ✅ Build successful.');
            }
          }
        }
      }
    } else if (lang === 'go') {
      const modPath = path.join(executionCwd, 'go.mod');
      if (await fs.pathExists(modPath)) {
        log('   Running go mod tidy...');
        const tidyRes = shell.exec('go mod tidy', { cwd: executionCwd, silent: true });
        if (tidyRes.code !== 0) {
          log(`   ❌ go mod tidy failed for '${id}'.`);
          log(tidyRes.stderr);
        } else {
          log('   ✅ Go modules tidied.');
        }
      }
    } else if (lang === 'python') {
      const pyProjectTomlPath = path.join(executionCwd, 'pyproject.toml');
      const requirementsTxtPath = path.join(executionCwd, 'requirements.txt');
      if (await fs.pathExists(pyProjectTomlPath)) {
        log('   Installing with pip from pyproject.toml...');
        const installRes = shell.exec('pip install --quiet .', {
          cwd: executionCwd,
          silent: true,
        });
        if (installRes.code !== 0) {
          log(`   ❌ pip install failed for '${id}'.\n${installRes.stderr}`);
        } else {
          log('   ✅ Python dependencies installed.');
        }
      } else if (await fs.pathExists(requirementsTxtPath)) {
        log('   Installing with pip from requirements.txt...');
        const installRes = shell.exec(`pip install --quiet -r "${requirementsTxtPath}"`, {
          cwd: executionCwd,
          silent: true,
        });
        if (installRes.code !== 0) {
          log(`   ❌ pip install failed for '${id}'.\n${installRes.stderr}`);
        } else {
          log('   ✅ Python dependencies installed.');
        }
      }
    } else if (lang === 'dart') {
      const pubspecPath = path.join(executionCwd, 'pubspec.yaml');
      if (await fs.pathExists(pubspecPath)) {
        log('   Running dart pub get...');
        const pubGetRes = shell.exec('dart pub get', { cwd: executionCwd, silent: true });
        if (pubGetRes.code !== 0) {
          log(`   ❌ dart pub get failed for '${id}'.`);
          log(pubGetRes.stderr);
        } else {
          log('   ✅ Dart dependencies installed.');
        }
      }
    } else {
      log(`   - No specific dependency installation for language: ${lang}`);
    }
    log(`✅ Finished dependency check for '${id}'.`);
  } catch (error) {
    log(`❌ Error installing dependencies for '${id}': ${error}`);
  }
}

module.exports = { installDependencies }; 