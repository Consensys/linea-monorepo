/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.zktracer.container.stacked.list;

import java.util.ArrayList;
import java.util.Collection;
import java.util.function.Predicate;
import java.util.function.UnaryOperator;

import net.consensys.linea.zktracer.container.ModuleOperation;

class CountedList<E extends ModuleOperation> extends ArrayList<E> {
  boolean countDirty = true;
  int count = 0;

  public CountedList() {
    super();
  }

  public CountedList(int initialCapacity) {
    super(initialCapacity);
  }

  @Override
  public boolean add(E e) {
    this.countDirty = true;
    return super.add(e);
  }

  @Override
  public E set(int index, E element) {
    this.countDirty = true;
    return super.set(index, element);
  }

  @Override
  public E remove(int index) {
    this.countDirty = true;
    return super.remove(index);
  }

  @Override
  public boolean remove(Object o) {
    this.countDirty = true;
    return super.remove(o);
  }

  @Override
  public boolean addAll(Collection<? extends E> c) {
    this.countDirty = true;
    return super.addAll(c);
  }

  @Override
  public boolean addAll(int index, Collection<? extends E> c) {
    this.countDirty = true;
    return super.addAll(index, c);
  }

  @Override
  protected void removeRange(int fromIndex, int toIndex) {
    this.countDirty = true;
    super.removeRange(fromIndex, toIndex);
  }

  @Override
  public boolean removeAll(Collection<?> c) {
    this.countDirty = true;
    return super.removeAll(c);
  }

  @Override
  public boolean retainAll(Collection<?> c) {
    this.countDirty = true;
    return super.retainAll(c);
  }

  @Override
  public boolean removeIf(Predicate<? super E> filter) {
    this.countDirty = true;
    return super.removeIf(filter);
  }

  @Override
  public void replaceAll(UnaryOperator<E> operator) {
    this.countDirty = true;
    super.replaceAll(operator);
  }

  public E getLast() {
    return this.get(this.size() - 1);
  }

  int lineCount() {
    if (this.countDirty) {
      this.count = 0;
      for (ModuleOperation op : this) {
        this.count += op.lineCount();
      }
      this.countDirty = false;
    }

    return this.count;
  }
}
