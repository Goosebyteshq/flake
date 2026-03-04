package com.example;

import org.junit.Test;

import static org.junit.Assert.assertEquals;

public class SampleTest {
    @Test
    public void testPass() {
        assertEquals(1, 1);
    }

    @Test
    public void testFail() {
        assertEquals(1, 2);
    }
}
