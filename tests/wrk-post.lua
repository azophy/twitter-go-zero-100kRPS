--adapted from https://gist.github.com/haggen/2fd643ea9a261fea2094?permalink_comment_id=4185532#gistcomment-4185532
math.randomseed(os.clock())
local charset = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
local wordset = {"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

function randomString(length)
	local ret = {}
	local r
	for i = 1, length do
		r = math.random(1, #charset)
		table.insert(ret, charset:sub(r, r))
	end
	return table.concat(ret)
end

function randomSentence(length)
	local ret = {}
	local r
	for i = 1, length do
		r = math.random(1, #wordset)
		table.insert(ret, wordset[r])
	end
	return table.concat(ret, " ")
end

local username=randomString(10)
local content=randomSentence(10)

wrk.method = "POST"
wrk.body   = string.format("username=%s&content=%s", username, content)
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
