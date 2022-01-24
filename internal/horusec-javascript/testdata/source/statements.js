function TryStatement() {
    try {
        const value = 'test'
        console.log(value)
    } catch (err) {
        console.error(err)
    } finally {
        let sum = 1 + 1
        console.log(sum)
    }
}
