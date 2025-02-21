/**
 * @see https://prettier.io/docs/en/configuration.html
 * @type {import("prettier").Config}
 */
const config = {
  trailingComma: "all",
  tabWidth: 2,
  semi: true,
  singleQuote: false,
  printWidth: 80,
  endOfLine: "lf",
  singleAttributePerLine: true,
  importOrder: [
    "^(react/(.*)$)|^(react$)",
    "^(next/(.*)$)|^(next$)",
    "<THIRD_PARTY_MODULES>",
    "",
    "^types$",
    "^@/types/(.*)$",
    "^@/config/(.*)$",
    "^@/lib/(.*)$",
    "^@/hooks/(.*)$",
    "^@/components/(.*)$",
    "^@/styles/(.*)$",
    "",
    "^[./]",
  ],
  importOrderParserPlugins: ["typescript", "tsx", "jsx", "decorators-legacy"],
  importOrderTypeScriptVersion: "5.2.2",
  plugins: ["@ianvs/prettier-plugin-sort-imports"],
};

export default config;
