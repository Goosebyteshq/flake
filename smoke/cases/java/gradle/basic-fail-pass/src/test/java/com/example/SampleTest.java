package com.example;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

class SampleTest {
    @Test
    void passCase() {
        assertEquals(1, 1);
    }

    @Test
    void failCase() {
        assertEquals(1, 2);
    }
}
