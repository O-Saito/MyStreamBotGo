
local accent_map = {
    ["á"] = "a",
    ["é"] = "e",
    ["í"] = "i",
    ["ó"] = "o",
    ["ú"] = "u",
    ["â"] = "a",
    ["ê"] = "e",
    ["ô"] = "o",
    ["à"] = "a",
    ["ã"] = "a",
    ["õ"] = "o",
    ["Á"] = "A",
    ["É"] = "E",
    ["Í"] = "I",
    ["Ó"] = "O",
    ["Ú"] = "U",
    ["Â"] = "A",
    ["Ê"] = "E",
    ["Ô"] = "O",
    ["À"] = "A",
    ["Ã"] = "A",
    ["Õ"] = "O"
};

local M = {}

function M.higieniza_string(str)
    local lowerStr = string.lower(str)
    for accent, char in pairs(accent_map) do
        lowerStr = string.gsub(lowerStr, accent, char)
    end
    local cleanedStr = string.gsub(lowerStr, '[^%a%d%s]', '')
    return cleanedStr
end

return M
