/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.testutils;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.util.Optional;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;

public class CodeDelegationTestUtils {

  public static CodeDelegation createDelegation(final Address authority, final Address target) {
    final CodeDelegation delegation = mock(CodeDelegation.class);
    when(delegation.authorizer()).thenReturn(Optional.of(authority));
    when(delegation.address()).thenReturn(target);
    return delegation;
  }
}
