const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');
const routerPath = path.join(root, 'service', 'api', 'api-handler.go');
const openapiPath = path.join(root, 'doc', 'api.yaml');

const httpMethods = new Set(['GET', 'POST', 'PUT', 'DELETE', 'PATCH']);

function normalizeRouterPath(routePath) {
  return routePath.replace(/:([A-Za-z][A-Za-z0-9_]*)/g, '{$1}');
}

function routeKey(method, routePath) {
  return `${method.toUpperCase()} ${routePath}`;
}

function extractRouterRoutes(content) {
  const routes = new Set();
  const routePattern = /rt\.router\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"/g;

  for (const match of content.matchAll(routePattern)) {
    routes.add(routeKey(match[1], normalizeRouterPath(match[2])));
  }

  return routes;
}

function extractOpenapiRoutes(content) {
  const routes = new Set();
  let currentPath = null;

  for (const line of content.split(/\r?\n/)) {
    const pathMatch = line.match(/^  (\/[^:]+):\s*$/);
    if (pathMatch) {
      currentPath = pathMatch[1];
      continue;
    }

    const methodMatch = line.match(/^    ([a-z]+):\s*$/);
    if (currentPath && methodMatch && httpMethods.has(methodMatch[1].toUpperCase())) {
      routes.add(routeKey(methodMatch[1], currentPath));
    }
  }

  return routes;
}

function difference(left, right) {
  return [...left].filter((item) => !right.has(item)).sort();
}

function checkRoutes(
  routerContent = fs.readFileSync(routerPath, 'utf8'),
  openapiContent = fs.readFileSync(openapiPath, 'utf8'),
) {
  const routerRoutes = extractRouterRoutes(routerContent);
  const openapiRoutes = extractOpenapiRoutes(openapiContent);

  return {
    routerRoutes,
    openapiRoutes,
    missingFromOpenapi: difference(routerRoutes, openapiRoutes),
    missingFromRouter: difference(openapiRoutes, routerRoutes),
  };
}

if (require.main === module) {
  const result = checkRoutes();
  const errors = [];

  if (result.missingFromOpenapi.length > 0) {
    errors.push('Routes registered in Go but missing from doc/api.yaml:');
    result.missingFromOpenapi.forEach((route) => errors.push(`  - ${route}`));
  }

  if (result.missingFromRouter.length > 0) {
    errors.push('Routes documented in doc/api.yaml but missing from the Go router:');
    result.missingFromRouter.forEach((route) => errors.push(`  - ${route}`));
  }

  if (errors.length > 0) {
    console.error('API route coverage check failed:');
    errors.forEach((error) => console.error(error));
    process.exit(1);
  }

  console.log(`API route coverage check passed (${result.routerRoutes.size} routes).`);
}

module.exports = {
  checkRoutes,
  extractOpenapiRoutes,
  extractRouterRoutes,
  normalizeRouterPath,
};
