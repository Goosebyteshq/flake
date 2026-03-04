using Xunit;

public class UnitTest1
{
    [Fact]
    public void AddsNumbers()
    {
        Assert.Equal(2, 1 + 1);
    }

    [Fact]
    public void DuplicateRejected()
    {
        Assert.Equal(2, 1);
    }
}
