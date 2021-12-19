function listSummary()
    version = "5.1"
    a = 1
    listDetail = function ()
        a = a+1
        io.write(string.format("Hello world, from %s, number %d !\n", version, a))
    end
    return listDetail
end

function testNest()
    detail = listSummary()
    detail()
end

testNest()