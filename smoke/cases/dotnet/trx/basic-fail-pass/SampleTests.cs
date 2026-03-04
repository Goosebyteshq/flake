using Xunit;

public class SampleTests
{
    [Fact]
    public void PassCase()
    {
        Assert.Equal(1, 1);
    }

    [Fact]
    public void FailCase()
    {
        Assert.Equal(1, 2);
    }
}
